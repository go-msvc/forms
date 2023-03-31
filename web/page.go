package main

import (
	"context"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/go-msvc/utils/ms"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type pageGetHandler func(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
)

type pagePostHandler func(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
	formData url.Values,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
)

type CtxDeviceID struct{}
type CtxEmail struct{}
type CtxTargetURL struct{}

// data given to the page template
type TmplData struct {
	Body   interface{} //depends on the page
	NavBar TmplNavBar
}
type TmplNavBar struct {
	//Items...
	//User  *TmplUser
	Email string
}

type TmplUser struct {
}

func open(getHdlr pageGetHandler, postHdlr pagePostHandler) http.HandlerFunc {
	return pageHandlerFunc(false, getHdlr, postHdlr)
}

func secure(getHdlr pageGetHandler, postHdlr pagePostHandler) http.HandlerFunc {
	return pageHandlerFunc(true, getHdlr, postHdlr)
}

func pageHandlerFunc(securePage bool, getHdlr pageGetHandler, postHdlr pagePostHandler) http.HandlerFunc {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		log.Debugf("HTTP %s %s", httpReq.Method, httpReq.URL.Path)
		var cookie *sessions.Session
		var err error
		var deviceID string
		var session *forms.Session

		//create ctx passed to functions
		//internal secure data is not stored in it - so called functions cannot access/manipulate it
		//such as internal session id
		ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*5)
		defer func() {
			log.Debugf("calling ctx cancelFunc()")
			cancelFunc()
		}()

		var tmpl *template.Template
		var data interface{}
		defer func() {
			if err != nil {
				if ec, ok := err.(errorWithCode); ok && ec.code != 0 {
					switch ec.Code() {
					case http.StatusTemporaryRedirect, http.StatusSeeOther:
						//update session before redirect
						if sessionErr := updSession(ctx, deviceID, session); sessionErr != nil {
							err = errors.Wrapf(sessionErr, "failed to write internal session")
						}
						//save cookie data
						delete(cookie.Values, "target-url")
						delete(cookie.Values, 42)
						delete(cookie.Values, "foo")
						for n, v := range cookie.Values {
							log.Debugf("  set cookie: %v = %v", n, v)
						}
						if cookieErr := cookie.Save(httpReq, httpRes); cookieErr != nil {
							err = errors.Wrapf(cookieErr, "failed to write cookie data")
						}
						//now ready to redirect
						httpReq.Method = http.MethodGet
						http.Redirect(httpRes, httpReq, ec.targetURL, ec.Code())
					default:
						log.Errorf("err=(%T)%+v", err, err)
						http.Error(httpRes, "unexpected error", http.StatusNotAcceptable)
					}
					return
				}
			}

			// update internal session (todo: only if modified)
			if err == nil && session != nil {
				if sessionErr := updSession(ctx, deviceID, session); sessionErr != nil {
					err = errors.Wrapf(sessionErr, "failed to write internal session")
				}
			}

			// write cookie data before writing content (todo: only if modified?)
			if err == nil {
				for n, v := range cookie.Values {
					log.Debugf("  set cookie: %v = %v", n, v)
				}
				if cookieErr := cookie.Save(httpReq, httpRes); cookieErr != nil {
					err = errors.Wrapf(cookieErr, "failed to write cookie data")
				}
			}

			if err == nil {
				if tmpl == nil {
					err = errors.Errorf("no page template to render")
				}
			}

			if err != nil {
				log.Errorf("Failed: %+v", err)
				showPage(ctx, errorTemplate, ErrorData{
					Message: fmt.Sprintf("Error: %+s", err),
				}, httpRes)
				return
			} else {
				//prepare template data to render this page
				tmplData := TmplData{
					Body: data,
				}

				//	"Nav":  map[string]interface{}{},

				if !session.Authenticated {
					// tmplData["User"] = TmplUser{
					// 	"Email": "",
					// }
				} else {
					tmplData.NavBar.Email = session.Email
				}
				log.Debugf("Rendering with tmplData:")
				//log.Debugf("  {{.NavBar.Items...}}: %+v", tmplData.NavBar)
				log.Debugf("  {{.NavBar}}: %+v", tmplData.NavBar)
				log.Debugf("  {{.Body}}:   (%T)%+v", tmplData.Body, tmplData.Body)

				httpRes.Header().Set("Content-Type", "text/html")
				if err = tmpl.ExecuteTemplate(httpRes, "page", tmplData); err != nil {
					log.Errorf("page template failed: %+v", err)
					err = errors.Wrapf(err, "failed to exec template")
					return
				}
			}
		}()

		//get cookie data from the browser - it always returns a value
		//this data is stored in the browser, encrypted, but still clever user
		//may figure out a way to manipulate this... so considered less secure
		//than internal session
		cookie, err = cookieStore.Get(httpReq, sessionAppName)
		if err != nil {
			err = errors.Wrapf(err, "failed to get cookies data")
			return
		}

		log.Debugf("CLIENT cookie:")
		for key, val := range cookie.Values {
			log.Debugf("  (%T)%+v : (%T)%+v", key, key, val, val)
		}

		if targetURL, ok := cookie.Values["target-url"]; ok && targetURL != "" {
			ctx = context.WithValue(ctx, CtxTargetURL{}, targetURL)
		}

		//when device id cannot be retrieved from cookie, assign a new device ID
		//this typically happens first time a device is used or after browser history was cleared
		//the device ID is internally associated with an session to avoid need to user to login repeatedly
		var ok bool
		deviceID, ok = cookie.Values["device-id"].(string)
		if !ok || deviceID == "" {
			deviceID = uuid.New().String()
			cookie.Values["device-id"] = deviceID
			log.Debugf("NEW device: %+v", deviceID)
		} else {
			log.Debugf("HAS device: %+v", deviceID)
		}
		ctx = context.WithValue(ctx, CtxDeviceID{}, deviceID)
		session, err = getSession(ctx, deviceID)
		if err != nil {
			//can't get an existing/new session
			//will also not be able to create one so no point to try login
			err = errors.Wrapf(err, "failed to get session")
			return
		}

		log.Debugf("secure=%v authenticated=%v", securePage, session.Authenticated)
		if securePage && !session.Authenticated {
			//user need to login to see this page
			//store requested URL in cookie then display login form
			log.Debugf("redirect to login...")
			cookie.Values["target-url"] = httpReq.URL.Path
			tmpl = loginEmailTemplate
			data = nil
			err = nil
			return
		}

		//proceed to non-secure page or securePage with authenticated
		//prepare page params
		params := map[string]string{}
		for n, v := range httpReq.URL.Query() {
			params[n] = fmt.Sprintf("%v", v)
			log.Debugf("param[%s]=\"%s\" (from URL Query)", n, params[n])
		}
		vars := mux.Vars(httpReq)
		for n, v := range vars {
			params[n] = v
			log.Debugf("param[%s]=\"%s\" (from URL vars)", n, params[n])
		}

		switch httpReq.Method {
		case http.MethodGet:
			tmpl, data, err = getHdlr(ctx, session, params)
			if err != nil {
				err = errors.Wrapf(err, "get handler failed")
				return
			}

		case http.MethodPost:
			if postHdlr == nil {
				err = errors.Errorf("POST not expected on this page")
				return
			}
			if err = httpReq.ParseForm(); err != nil {
				err = errors.Wrapf(err, "failed to parse the form data")
				return
			}
			log.Debugf("form data: %+v", httpReq.PostForm)
			tmpl, data, err = postHdlr(ctx, session, params, httpReq.PostForm)
			if err != nil {
				//also for redirect...
				log.Debugf("postHandler failed: (%T)%+v", err, err)
				return
			}
			log.Debugf("postHandler succeeded")

		default:
			err = errors.Errorf("method not supported")
		}
	} //handlerFunc()
} //pageHandlerFunc()

func getSession(ctx context.Context, deviceID string) (*forms.Session, error) {
	//lookup session for this device ID
	//if none - create a temp session and redirect to login
	//get unique internal session identified by device ID
	log.Debugf("getSession(deviceID:%s)", deviceID)
	res, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_session",
		},
		formsTTL,
		formsinterface.GetSessionRequest{
			DeviceID: deviceID,
		},
		formsinterface.GetSessionResponse{})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get session(device_id:%s)", deviceID)
	}

	//got a session
	var session *forms.Session
	{
		s := res.(formsinterface.GetSessionResponse).Session
		session = &s
	}
	if session.Data == nil {
		session.Data = map[string]interface{}{}
	}
	log.Debugf("device(%s).session(%s).(auth:%v,email:%v) retrieved with %d values:", deviceID, session.ID, session.Authenticated, session.Email, len(session.Data))
	for n, v := range session.Data {
		log.Debugf("  session[\"%s\"] = (%T)%v", n, v, v)
	}
	return session, nil
} //getSession()

func updSession(ctx context.Context, deviceID string, session *forms.Session) error {
	log.Debugf("device(%s).session(%s).(auth:%v,email:%v) updating with %d values:", deviceID, session.ID, session.Authenticated, session.Email, len(session.Data))
	for n, v := range session.Data {
		log.Debugf("  session[\"%s\"] = (%T)%v", n, v, v)
	}
	_, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "upd_session",
		},
		formsTTL,
		formsinterface.UpdSessionRequest{
			DeviceID: deviceID,
			Session:  *session,
		},
		nil)
	if err != nil {
		log.Errorf("failed to upd session.id(%s): %+v", session.ID, err)
		return errors.Wrapf(err, "failed to update session(%s)", session.ID)
	} else {
		log.Debugf("Updated session")
	}
	return nil
} //updSession()

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
		formsTTL,
		formsinterface.AddDocRequest{Doc: doc},
		formsinterface.AddDocResponse{})
	if err != nil {
		return forms.Doc{}, errors.Wrapf(err, "failed to create document")
	}
	return res.(formsinterface.AddDocResponse).Doc, nil
}

func renderPage(w io.Writer, t *template.Template, data any) error {
	if err := t.ExecuteTemplate(w, "page", data); err != nil {
		return errors.Wrapf(err, "failed to exec template")
	}
	return nil
}

type errorWithCode struct {
	error
	code      int
	targetURL string
}

// func (e errorWithCode) Error() string {
// 	return e.error.Error()
// }

func (e errorWithCode) Code() int {
	return e.code
}

func ErrorRedirect(targetURL string) errorWithCode {
	return errorWithCode{
		error:     errors.Errorf("redirect to %s", targetURL),
		code:      http.StatusSeeOther, //http.StatusTemporaryRedirect = redirect with same method e.g. POST
		targetURL: targetURL,
	}
}
