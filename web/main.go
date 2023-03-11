package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/go-msvc/logger"
	"github.com/go-msvc/nats-utils"
	"github.com/go-msvc/utils/ms"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"

	"github.com/gomarkdown/markdown"
)

var (
	log         = logger.New().WithLevel(logger.LevelDebug)
	msClient    ms.Client
	redisClient *redis.Client
	formsDomain = "forms"
	formsTTL    = time.Second * 1
)

func main() {
	//preload some templates
	loadResources()

	//routers directs http to the correct page from the URL path
	r := mux.NewRouter()
	r.HandleFunc("/home", page("home"))
	r.HandleFunc("/login", page("login"))
	r.HandleFunc("/campaign/{id}", page2(showCampaign, postCampaign))
	r.HandleFunc("/", page("home")) //defaultHandler)
	http.Handle("/", r)

	//fileServer serves static files such as style sheets from the ./resources folder
	//note: templates
	fileServer := http.FileServer(http.Dir("./resources"))
	http.Handle("/resources/", httpLogger(http.StripPrefix("/resources", fileServer)))

	//ms client
	msClientConfig := nats.ClientConfig{
		Config: nats.Config{
			Domain: "forms-web",
		},
	}
	if err := msClientConfig.Validate(); err != nil {
		panic(fmt.Sprintf("client config: %+v", err))
	}
	var err error
	msClient, err = msClientConfig.Create()
	if err != nil {
		panic(fmt.Sprintf("failed to create ms client: %+v", err))
	}

	//redis client
	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	if err != nil {
		panic(fmt.Sprintf("failed to create redis client: %+v", err))
	}

	//start the web server
	http.ListenAndServe("localhost:8080", nil)
}

func httpLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTTP %s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

// var pageTemplate *template.Template
var (
	pageTemplate              map[string]*template.Template
	formTemplate              *template.Template
	formSubmittedTemplate     *template.Template
	campaignSubmittedTemplate *template.Template
	errorTemplate             *template.Template
)

func loadResources() {
	pageTemplate = make(map[string]*template.Template)
	pageTemplate["home"] = loadTemplates([]string{"home", "page"})
	pageTemplate["login"] = loadTemplates([]string{"login", "page"})

	formTemplate = loadTemplates([]string{"form", "page"})
	formSubmittedTemplate = loadTemplates([]string{"form-submitted", "page"})
	campaignSubmittedTemplate = loadTemplates([]string{"campaign-submitted", "page"})
	errorTemplate = loadTemplates([]string{"error", "page"})
}

func loadTemplates(templateNames []string) *template.Template {
	templateFileNames := []string{}
	for _, n := range templateNames {
		templateFileNames = append(templateFileNames, "./templates/"+n+".tmpl")
	}
	t, err := template.ParseFiles(templateFileNames...)
	if err != nil {
		panic(fmt.Sprintf("failed to load template files: %v: %+v", templateFileNames, err))
	}
	log.Debugf("loaded %v", templateFileNames)
	return t
}

func page(pageName string) http.HandlerFunc {
	pt, ok := pageTemplate[pageName]
	if !ok {
		panic(fmt.Sprintf("missing page(%s) template", pageName))
	}
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		log.Debugf("HTTP %s %s", httpReq.Method, httpReq.URL.Path)
		var clientSession *sessions.Session
		var err error
		// var internalSession internal.Session
		defer func() {
			log.Errorf("err=%+v", err)
			//update internal session (todo: only if modified)
			// if sessionErr := updateInternalSession(internalSession); sessionErr != nil {
			// 	err = errors.Wrapf(sessionErr, "failed to write internal session")
			// }
			// write cookie data before writing content (todo: only if modified?)
			// clientSession.Values["internal-session-id"] = internalSession.ID
			if cookieErr := clientSession.Save(httpReq, httpRes); cookieErr != nil {
				err = errors.Wrapf(cookieErr, "failed to write cookie data")
			}

			if err != nil {
				log.Errorf("Failed: %+v", err)
				http.Error(httpRes, fmt.Sprintf("failed: %v", err), http.StatusInternalServerError)
				return
			}
		}()

		//get cookie data from the browser - it always returns a value
		//this data is stored in the browser, encrypted, but still clever user
		//may figure out a way to manipulate this... so considered less secure
		//than internal session
		clientSession, err = cookieStore.Get(httpReq, sessionAppName)
		if err != nil {
			err = errors.Wrapf(err, "failed to get cookies data")
			return
		}

		log.Debugf("retrieved %d cookies:", len(clientSession.Values))
		for key, val := range clientSession.Values {
			log.Debugf("  (%T)%+v : (%T)%+v", key, key, val, val)
		}

		//create ctx passed to functions
		//internal secure data is not stored in it - so called functions cannot access/manipulate it
		//such as internal session id
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*5)
		defer func() {
			log.Debugf("calling ctx cancelFunc()")
			cancelFunc()
		}()

		//todo: check auth and redirect to login, but allow to come back to same page after login...
		if userID := clientSession.Values["user-id"]; userID != "" {
			// userToken := internalSession.Data["user-token"]
			// log.Debugf("todo: userToken(%s)", userToken)
			// if userToken != "" {
			// 	user, err = auth.GetUser(userToken)
			// 	if err != nil {
			// 		err = errors.Wrapf(err, "failed to get user info")
			// 		return
			// 	}
			// }
		}

		//todo: if not authenticated and requested auth page, store request and then redirect to login
		//...

		switch httpReq.Method {
		case http.MethodGet:
			showPage(ctx, pt, nil, httpRes)
		case http.MethodPost:
			if err = httpReq.ParseForm(); err != nil {
				err = errors.Wrapf(err, "failed to parse the form data")
				return
			}
			log.Debugf("form data: %+v", httpReq.PostForm)
		default:
			err = errors.Errorf("method not supported")
		}
	}
} //page()

func showPage(ctx context.Context, t *template.Template, data any, httpRes http.ResponseWriter) {
	httpRes.Header().Set("Content-Type", "text/html")
	log.Debugf("t=%+v", t)
	if err := renderPage(httpRes, t, data); err != nil {
		http.Error(httpRes, fmt.Sprintf("failed: %+v", err), http.StatusInternalServerError)
		return
	}
}

func postForm(ctx context.Context, session *forms.Session, values url.Values) (forms.Doc, error) {
	//log.Debugf("submitForm: %+v", values)

	//todo: security should include doc token fetched from db, to restrict update

	//get form id and revision
	formID := session.Data["form_id"].(string)
	formRevValue, gotRev := session.Data["form_rev"]
	if formID == "" || !gotRev {
		log.Errorf("1")
		return forms.Doc{}, errors.Errorf("missing form id/rev")
	}
	formRev, err := strconv.ParseInt(fmt.Sprintf("%v", formRevValue), 10, 64)
	if err != nil {
		log.Errorf("1")
		return forms.Doc{}, errors.Wrapf(err, "invalid form rev(%v)", formRevValue)
	}
	//log.Debugf("submit form.id(%s).rev(%v)", formID, formRev)
	//get doc id (if editing existing doc)
	// docID := values["doc_id"]
	// docRev := values["doc_rev"]
	//...todo: if defined - do upd_doc instead of add_doc

	doc := forms.Doc{
		FormID:  formID,
		FormRev: int(formRev),
		//ID:      docID,
		Data: map[string]interface{}{},
	}
	for n, v := range values {
		doc.Data[n] = v
	}

	//use ms client to store the document
	res, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "add_doc",
		},
		time.Millisecond*time.Duration(formsTTL),
		formsinterface.AddDocRequest{Doc: doc},
		formsinterface.AddDocResponse{})
	if err != nil {
		return forms.Doc{}, errors.Wrapf(err, "failed to create document")
	}
	return res.(formsinterface.AddDocResponse).Doc, nil
}

func defaultHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "text/html")
	if err := renderPage(httpRes, pageTemplate["home"], nil); err != nil {
		log.Errorf("failed to render: %+v", err)
		http.Error(httpRes, fmt.Sprintf("failed: %+v", err), http.StatusInternalServerError)
		return
	}
}

func renderPage(w io.Writer, t *template.Template, data any) error {
	if err := t.ExecuteTemplate(w, "page", data); err != nil {
		return errors.Wrapf(err, "failed to exec template")
	}
	return nil
}

// Note: Don't store your key in your source code. Pass it via an
// environmental variable, or flag (or both), and don't accidentally commit it
// alongside your code. Ensure your key is sufficiently random - i.e. use Go's
// crypto/rand or securecookie.GenerateRandomKey(32) and persist the result.
// Ensure SESSION_KEY exists in the environment, or sessions will fail.
var (
	sessionAppName   string
	secretSessionKey []byte
	cookieStore      sessions.Store
)

func init() {
	sessionAppName = os.Getenv("SESSION_APP_NAME")
	if sessionAppName == "" {
		sessionAppName = "noname-app"
	}
	keyString := os.Getenv("SESSION_SECRET_KEY")
	if keyString == "" {
		keyString = "default-insecure-key"
	}
	secretSessionKey = []byte(keyString)
	cookieStore = sessions.NewCookieStore([]byte(secretSessionKey))
}

var apiURL = "http://localhost:12345"

type ErrorData struct {
	Message string
}

func renderHeaderHTML(h forms.Header) forms.Header {
	//generate HTML descriptions
	// ext := markdownparser.CommonExtensions | markdownparser.AutoHeadingIDs
	// mdp := markdownparser.NewWithExtensions(ext)
	h.HtmlTitle = template.HTML(markdown.ToHTML([]byte(h.Title), nil, nil))
	if h.Description != "" {
		h.HtmlDescription = template.HTML(markdown.ToHTML([]byte(h.Description), nil, nil))
	} else {
		h.Description = ""
	}
	return h
}
