package main

import (
	"context"
	"encoding/json"
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
	r.HandleFunc("/form/{id}", formHandler)
	r.HandleFunc("/campaign/{id}", campaignHandler)
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
	pageTemplate         map[string]*template.Template
	formTemplate         *template.Template
	submittedDocTemplate *template.Template
	errorTemplate        *template.Template
)

func loadResources() {
	pageTemplate = make(map[string]*template.Template)
	pageTemplate["home"] = loadTemplates([]string{"home", "page"})
	pageTemplate["login"] = loadTemplates([]string{"login", "page"})

	formTemplate = loadTemplates([]string{"form", "page"})
	submittedDocTemplate = loadTemplates([]string{"submitted", "page"})
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
			showPage(ctx, pt, nil, httpRes)
		case http.MethodPost:
			if err = httpReq.ParseForm(); err != nil {
				err = errors.Wrapf(err, "failed to parse the form data")
				return
			}
			log.Debugf("form data: %+v", httpReq.PostForm)
			// if doc, err := postForm(ctx, httpReq.PostForm); err != nil {
			// 	log.Errorf("failed to post: %+v", err)
			// 	err = errors.Wrapf(err, "failed to submit the form data")
			// 	return
			// } else {
			// 	//show details of submitted documents
			// 	log.Debugf("Submitted: %+v", doc)
			// 	docData := map[string]interface{}{
			// 		"ID":        doc.ID,
			// 		"Rev":       doc.Rev,
			// 		"Timestamp": doc.Timestamp,
			// 		"FormID":    doc.FormID,
			// 	}
			// 	showPage(ctx, submittedDocTemplate, docData, httpRes)
			// }
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

func postForm(ctx context.Context, values url.Values) (forms.Doc, error) {
	log.Debugf("submitForm: %+v", values)

	//todo: security should include doc token fetched from db, to restrict update

	//get form id and revision
	formID := values.Get("form_id")
	formRevStr := values.Get("form_rev")
	if formID == "" || formRevStr == "" {
		log.Errorf("1")
		return forms.Doc{}, errors.Errorf("missing form id/rev")
	}
	formRev, err := strconv.ParseInt(formRevStr, 10, 64)
	if err != nil {
		log.Errorf("1")
		return forms.Doc{}, errors.Wrapf(err, "invalid form rev(%s)", formRevStr)
	}

	log.Debugf("submit form.id(%s).rev(%v)", formID, formRev)
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
		//use switch to skip the fields that are not part of the document
		//todo: remove later when these are managed in context
		switch n {
		case "form_id":
		case "form_rev":
		case "doc_id":
		case "doc_rev":
		default:
			doc.Data[n] = v
		}
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

type ErrorData struct {
	Message string
}

//	func updateInternalSession(s internal.Session) error {
//		url := apiURL + "/sessions/" + s.ID
//		jsonSession, err := json.Marshal(s)
//		if err != nil {
//			return errors.Wrapf(err, "failed to encode internal session")
//		}
//		httpReq, _ := http.NewRequest(http.MethodPut, url, bytes.NewReader(jsonSession))
//		httpReq.Header.Set("Content-Type", "application/json")
//		if _, err = http.DefaultClient.Do(httpReq); err != nil {
//			return errors.Wrapf(err, "failed to update internal session")
//		}
//		return nil
//	}
func formHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	vars := mux.Vars(httpReq)
	id := vars["id"]

	//use ms client to fetch the form
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

	switch httpReq.Method {
	case http.MethodGet:
		form.Header = renderHeaderHTML(form.Header)
		for i, s := range form.Sections {
			s.Header = renderHeaderHTML(s.Header)
			for itemIndex, item := range s.Items {
				if item.Header != nil {
					*item.Header = renderHeaderHTML(*item.Header)
				}
				if item.Field != nil {
					item.Field.Header = renderHeaderHTML(item.Field.Header)
				}
				if item.Image != nil {
					item.Image.Header = renderHeaderHTML(item.Image.Header)
				}
				if item.Table != nil {
					item.Table.Header = renderHeaderHTML(item.Table.Header)
				}
				if item.Sub != nil {
					item.Sub.Header = renderHeaderHTML(item.Sub.Header)
				}
				s.Items[itemIndex] = item
			} //for each item
			form.Sections[i] = s
		} //for each section
		form.Action = fmt.Sprintf("/form/%s", form.ID)
		showForm(form, httpRes)
	case http.MethodPost:
		if err := httpReq.ParseForm(); err != nil {
			err = errors.Wrapf(err, "failed to parse the form data")
			return
		}
		log.Debugf("form data: %+v", httpReq.PostForm)
		if doc, err := postForm(context.Background() /*TODO*/, httpReq.PostForm); err != nil {
			log.Errorf("failed to post: %+v", err)
			err = errors.Wrapf(err, "failed to submit the form data")
			showPage(ctx, errorTemplate, ErrorData{
				Message: fmt.Sprintf("Failed to submit the document: %+s", err),
			}, httpRes)
			return
		} else {
			//show details of submitted documents
			log.Debugf("Submitted: %+v", doc)
			showPage(ctx, submittedDocTemplate, doc, httpRes)
		}
	default:
		http.Error(httpRes, "Method not allowed", http.StatusMethodNotAllowed)
	}
} //formHandler()

func campaignHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	vars := mux.Vars(httpReq)
	id := vars["id"]
	ctx := context.Background()
	res, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_campaign",
		},
		time.Millisecond*time.Duration(formsTTL),
		formsinterface.GetCampaignRequest{
			ID: id,
		},
		formsinterface.GetCampaignResponse{})
	if err != nil {
		log.Errorf("campaign.id(%s) not found", id)
		httpRes.Header().Set("Content-Type", "text/plain")
		http.Error(httpRes, fmt.Sprintf("unknown campaign id(%s)", id), http.StatusNotFound)
		return
	}
	campaign := res.(formsinterface.GetCampaignResponse).Campaign

	res, err = msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_form",
		},
		time.Millisecond*time.Duration(formsTTL),
		formsinterface.GetFormRequest{
			ID: campaign.FormID,
		},
		formsinterface.GetFormResponse{})
	if err != nil {
		log.Errorf("campaign(%s).form(%s) not found", campaign.ID, campaign.FormID)
		httpRes.Header().Set("Content-Type", "text/plain")
		http.Error(httpRes, fmt.Sprintf("unknown form id(%s)", id), http.StatusNotFound)
		return
	}
	form := res.(formsinterface.GetFormResponse).Form
	if campaign.StartTime.After(time.Now()) {
		log.Errorf("campaign(%s) only starts at %v", campaign.ID, campaign.StartTime)
		httpRes.Header().Set("Content-Type", "text/plain")
		http.Error(httpRes, fmt.Sprintf("campaign(%s) only starts at %v", campaign.ID, campaign.StartTime), http.StatusNotFound)
		return
	}
	if campaign.EndTime.Before(time.Now()) {
		log.Errorf("campaign(%s) ended at %v", campaign.ID, campaign.EndTime)
		httpRes.Header().Set("Content-Type", "text/plain")
		http.Error(httpRes, fmt.Sprintf("campaign(%s) ended at %v", campaign.ID, campaign.EndTime), http.StatusNotFound)
		return
	}

	switch httpReq.Method {
	case http.MethodGet:
		form.Header = renderHeaderHTML(form.Header)
		for i, s := range form.Sections {
			s.Header = renderHeaderHTML(s.Header)
			for itemIndex, item := range s.Items {
				if item.Header != nil {
					*item.Header = renderHeaderHTML(*item.Header)
				}
				if item.Field != nil {
					item.Field.Header = renderHeaderHTML(item.Field.Header)
				}
				if item.Image != nil {
					item.Image.Header = renderHeaderHTML(item.Image.Header)
				}
				if item.Table != nil {
					item.Table.Header = renderHeaderHTML(item.Table.Header)
				}
				if item.Sub != nil {
					item.Sub.Header = renderHeaderHTML(item.Sub.Header)
				}
				s.Items[itemIndex] = item
			} //for each item
			form.Sections[i] = s
		} //for each section
		form.Action = fmt.Sprintf("/campaign/%s", campaign.ID)
		showForm(form, httpRes)
	case http.MethodPost:
		if err := httpReq.ParseForm(); err != nil {
			err = errors.Wrapf(err, "failed to parse the form data")
			return
		}
		log.Debugf("form data: %+v", httpReq.PostForm)
		if doc, err := postForm(context.Background() /*TODO*/, httpReq.PostForm); err != nil {
			log.Errorf("failed to post: %+v", err)
			err = errors.Wrapf(err, "failed to submit the form data")
			showPage(ctx, errorTemplate, ErrorData{
				Message: fmt.Sprintf("Failed to submit the document: %+s", err),
			}, httpRes)
			return
		} else {
			log.Debugf("Submitted: %+v", doc)

			//send campaign notification
			notification := formsinterface.CampaignNotification{
				CampaingID: campaign.ID,
				DocID:      doc.ID,
			}
			log.Errorf("NOT YET SENDING Notification to %s: %+v", campaign.ID, notification)
			jsonNotification, _ := json.Marshal(notification)
			if i64, err := redisClient.LPush(ctx, campaign.ID, jsonNotification).Result(); err != nil {
				err = errors.Wrapf(err, "failed to send for processing")
				log.Errorf("failed: %+v", err)
				showPage(ctx, errorTemplate, ErrorData{
					Message: fmt.Sprintf("Failed to send for processing: %+s", err),
				}, httpRes)
				return
			} else {
				log.Debugf("Pushed notification result %d", i64)
			}

			//show details of submitted documents
			showPage(ctx, submittedDocTemplate, doc, httpRes)
		}
	default:
		http.Error(httpRes, "Method not allowed", http.StatusMethodNotAllowed)
	}
} //campaignHandler()

func showForm(form forms.Form, httpRes http.ResponseWriter) {
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
