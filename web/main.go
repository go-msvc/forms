package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/go-msvc/logger"
	"github.com/go-msvc/nats-utils"
	"github.com/go-msvc/utils/ms"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

var (
	log         = logger.New().WithLevel(logger.LevelDebug)
	msClient    ms.Client
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
	r.HandleFunc("/form/{id}", formHandler) //page("form"))
	r.HandleFunc("/", page("home"))         //defaultHandler)
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
	pageTemplate map[string]*template.Template
	formTemplate *template.Template
)

func loadResources() {
	pageTemplate = make(map[string]*template.Template)
	pageTemplate["home"] = loadTemplates([]string{"home", "page"})
	pageTemplate["login"] = loadTemplates([]string{"login", "page"})

	formTemplate = loadTemplates([]string{"form", "page"})
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

		//get internal session identified by client session
		// if internalSessionID, ok := clientSession.Values["internal-session-id"].(string); ok && internalSessionID != "" {
		// 	log.Debugf("internal-session-id: \"%s\"", internalSessionID)
		// 	if internalSession, err = getInternalSession(internalSessionID); err != nil {
		// 		err = errors.Wrapf(err, "failed to get internal-session")
		// 		return
		// 	} else {
		// 		log.Debugf("got internal-session(%s): %+v", internalSessionID, internalSession)
		// 	}
		// }
		// if internalSession.Data == nil {
		// 	internalSession.Data = map[string]interface{}{}
		// }

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
			log.Debugf("rendering %s ...", pageName)
			showPage(ctx, pt, httpRes)
			log.Debugf("rendered %s ...", pageName)
		case http.MethodPost:
			if err = httpReq.ParseForm(); err != nil {
				err = errors.Wrapf(err, "failed to parse the form data")
				return
			}
			log.Debugf("form data: %+v", httpReq.PostForm)
			postForm(ctx, httpReq.PostForm)
		default:
			err = errors.Errorf("method not supported")
		}
	}
} //page()

func showPage(ctx context.Context, t *template.Template, httpRes http.ResponseWriter) {
	httpRes.Header().Set("Content-Type", "text/html")
	log.Debugf("t=%+v", t)
	if err := renderPage(httpRes, t, nil); err != nil {
		http.Error(httpRes, fmt.Sprintf("failed: %+v", err), http.StatusInternalServerError)
		return
	}
}

func postForm(ctx context.Context, values url.Values) {
	//values := map[string]interface{}{} //todo... get from ?

	log.Debugf("submitForm: %+v", values)
}

func defaultHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "text/html")
	if err := renderPage(httpRes, pageTemplate["home"], nil); err != nil {
		log.Errorf("failed to render: %+v", err)
		http.Error(httpRes, fmt.Sprintf("failed: %+v", err), http.StatusInternalServerError)
		return
	}
}

func renderPage(w io.Writer, t *template.Template, data map[string]interface{}) error {
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

// func getInternalSession(id string) (internal.Session, error) {
// 	url := apiURL + "/sessions"
// 	if id != "" {
// 		url += "/" + id
// 	}
// 	httpReq, _ := http.NewRequest(http.MethodGet, url, nil)
// 	httpRes, err := http.DefaultClient.Do(httpReq)
// 	if err != nil {
// 		return internal.Session{}, errors.Wrapf(err, "failed to get internal session")

// 		//todo: if expired - need to create a new session, which will auto prompt for login
// 		//store id in client session, so not have to ask it again...

// 	}
// 	var s internal.Session
// 	if err := json.NewDecoder(httpRes.Body).Decode(&s); err != nil {
// 		return internal.Session{}, errors.Wrapf(err, "failed to decode internal session")
// 	}
// 	return s, nil
// }

// func updateInternalSession(s internal.Session) error {
// 	url := apiURL + "/sessions/" + s.ID
// 	jsonSession, err := json.Marshal(s)
// 	if err != nil {
// 		return errors.Wrapf(err, "failed to encode internal session")
// 	}
// 	httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonSession))
// 	httpReq.Header.Set("Content-Type", "application/json")
// 	if _, err = http.DefaultClient.Do(httpReq); err != nil {
// 		return errors.Wrapf(err, "failed to update internal session")
// 	}
// 	return nil
// }

var (
	testForms = map[string]forms.Form{
		"1": {
			ID:     "1",
			Header: forms.Header{Title: "FormTitle1", Description: "FormDesc1"},
			Sections: []forms.Section{
				{
					FirstSection: true,
					Name:         "sec_1",
					Header:       forms.Header{Title: "Section 1", Description: "This is section 1..."},
					Items: []forms.Item{
						{Field: &forms.Field{Header: forms.Header{Title: "Field1", Description: "Short text entry"}, Name: "field_1", Short: &forms.Short{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field2", Description: "Another short text entry"}, Name: "field_2", Short: &forms.Short{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field3", Description: "FieldDesc2"}, Name: "field_3", Integer: &forms.Integer{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field4", Description: "FieldDesc2"}, Name: "field_4", Number: &forms.Number{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field5", Description: "FieldDesc2"}, Name: "field_5", Choice: &forms.Choice{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field6", Description: "FieldDesc2"}, Name: "field_6", Selection: &forms.Selection{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field7", Description: "FieldDesc2"}, Name: "field_7", Text: &forms.Text{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field8", Description: "FieldDesc2"}, Name: "field_8", Date: &forms.Date{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field9", Description: "FieldDesc2"}, Name: "field_9", Time: &forms.Time{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field10", Description: "FieldDesc2"}, Name: "field_10", Duration: &forms.Duration{}}},
						{Header: &forms.Header{Title: "header line", Description: "with a short description..."}},
						{Image: &forms.Image{}},
						{Table: &forms.Table{}},
						{Sub: &forms.Sub{}},
					},
				},
				{
					Name:   "sec_2",
					Header: forms.Header{Title: "Section 2", Description: "This is section 2..."},
					Items: []forms.Item{
						{Field: &forms.Field{Header: forms.Header{Title: "Field1", Description: "FieldDesc1"}, Name: "field_1", Short: &forms.Short{}}},
						{Field: &forms.Field{Header: forms.Header{Title: "Field2", Description: "FieldDesc2"}, Name: "field_2", Short: &forms.Short{}}},
					},
				},
			},
		},
	}
)

type formTemplateData struct {
	Title       string
	Description string
	Sections    []formTemplateSection
}

type formTemplateSection struct {
	Title       string
	Description string
	Items       []formTemplateItem
}

type formTemplateItem struct {
}

func formHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	vars := mux.Vars(httpReq)
	id := vars["id"]

	//us ms client to fetch the form
	//ms-client use one id for context, request and own domain, as it does only one request then terminates
	ctx := context.Background()
	res, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_form",
		},
		time.Millisecond*time.Duration(formsTTL),
		formsinterface.GetFormRequest{
			ID: id,
		},
		formsinterface.GetFormResponse{})
	if err != nil {
		log.Errorf("form.id(%s) not found", id)
		httpRes.Header().Set("Content-Type", "text/plain")
		http.Error(httpRes, fmt.Sprintf("unknown form id(%s)", id), http.StatusNotFound)
		return
	}
	log.Debugf("Got res (%T)%+v", res, res)
	form := res.(formsinterface.GetFormResponse).Form

	// form, ok := testForms[id]
	// if !ok {
	// 	log.Errorf("form.id(%s) not found", id)
	// 	httpRes.Header().Set("Content-Type", "text/plain")
	// 	http.Error(httpRes, fmt.Sprintf("unknown form id(%s)", id), http.StatusNotFound)
	// 	return
	// }

	switch httpReq.Method {
	case http.MethodGet:
		showForm(form, httpRes)
	case http.MethodPost:
		if err := httpReq.ParseForm(); err != nil {
			err = errors.Wrapf(err, "failed to parse the form data")
			return
		}
		log.Debugf("form data: %+v", httpReq.PostForm)
		postForm(context.Background() /*TODO*/, httpReq.PostForm)
	default:
		http.Error(httpRes, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func showForm(form forms.Form, httpRes http.ResponseWriter) { //httpRes http.ResponseWriter, httpReq *http.Request)
	//prepare data used by the template to render the form
	// formData := formTemplateData{
	// 	Title:       form.Header.Title,
	// 	Description: form.Header.Description,
	// 	Sections:    []formTemplateSection{},
	// }
	// for _, s := range form.Sections {
	// 	fs := formTemplateSection{
	// 		Title:       s.Header.Title,
	// 		Description: s.Header.Description,
	// 	}
	// 	formData.Sections = append(formData.Sections, fs)
	// }

	//load template at runtime while designing...
	//later comment out and use preloaded one only
	formTemplate := loadTemplates([]string{"form", "page"})

	//render the form into HTML and javascript
	if err := formTemplate.ExecuteTemplate(httpRes, "page", form /*formData*/); err != nil {
		log.Errorf("form(%s) rendering failed: %+v", form.ID, err)
		httpRes.Header().Set("Content-Type", "text/plain")
		http.Error(httpRes, fmt.Sprintf("form(%s) rendering failed: %+v", form.ID, err), http.StatusNotFound)
		return
	}
}
