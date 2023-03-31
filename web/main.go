package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"time"

	"github.com/go-msvc/forms"
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
	r.HandleFunc("/home", open(page(homeTemplate), nil))
	r.HandleFunc("/login", open(page(loginEmailTemplate), loginEmailHandler))
	r.HandleFunc("/otp", open(page(loginOtpTemplate), loginOtpHandler))
	r.HandleFunc("/logout", open(logoutHandler, nil))
	r.HandleFunc("/user", secure(userHomeGetHandler, nil))
	r.HandleFunc("/user/campaign/{campaign_id}", secure(myCampaign, nil)) //for submission
	r.HandleFunc("/campaign/{id}", secure(showCampaign, postCampaign))    //for submission
	r.HandleFunc("/", secure(page(homeTemplate), nil))                    //defaultHandler)
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

	//start the web server
	http.ListenAndServe(":8080", nil)
}

func httpLogger(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debugf("HTTP %s %s", r.Method, r.URL.Path)
		h.ServeHTTP(w, r)
	})
}

var (
	homeTemplate              *template.Template
	loginEmailTemplate        *template.Template
	loginOtpTemplate          *template.Template
	userHomeTemplate          *template.Template
	userCampaignTemplate      *template.Template
	formTemplate              *template.Template
	formSubmittedTemplate     *template.Template
	campaignSubmittedTemplate *template.Template
	errorTemplate             *template.Template
)

func loadResources() {
	homeTemplate = loadTemplates([]string{"home", "page"})
	loginEmailTemplate = loadTemplates([]string{"login-email-form", "page"})
	loginOtpTemplate = loadTemplates([]string{"login-otp-form", "page"})
	userHomeTemplate = loadTemplates([]string{"user-home", "page"})
	userCampaignTemplate = loadTemplates([]string{"user-campaign", "page"})
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

func page(pt *template.Template) pageGetHandler {
	//here we can fail on startup if there are missing templates
	if pt == nil {
		panic("missing template")
	}

	return func(
		ctx context.Context,
		session *forms.Session,
		params map[string]string,
	) (
		tmpl *template.Template,
		tmplData interface{},
		err error,
	) {
		log.Debugf("showPage(%+v)", params)
		pageData := map[string]interface{}{}
		return pt, pageData, nil
	}
} //page()

func defaultHandler(httpRes http.ResponseWriter, httpReq *http.Request) {
	httpRes.Header().Set("Content-Type", "text/html")
	if err := renderPage(httpRes, homeTemplate, nil); err != nil {
		log.Errorf("failed to render: %+v", err)
		http.Error(httpRes, fmt.Sprintf("failed: %+v", err), http.StatusInternalServerError)
		return
	}
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
