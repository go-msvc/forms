package main

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/go-msvc/utils/ms"
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

func page2(getHdlr pageGetHandler, postHdlr pagePostHandler) http.HandlerFunc {
	return func(httpRes http.ResponseWriter, httpReq *http.Request) {
		log.Debugf("HTTP %s %s", httpReq.Method, httpReq.URL.Path)
		var clientSession *sessions.Session
		var err error
		var internalSession forms.Session

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
				log.Errorf("err=%+v", err)
			}

			// update internal session (todo: only if modified)
			if err == nil {
				if sessionErr := updateInternalSession(ctx, &internalSession); sessionErr != nil {
					err = errors.Wrapf(sessionErr, "failed to write internal session")
				}
			}

			// write cookie data before writing content (todo: only if modified?)
			if err == nil {
				//log.Debugf("Storing internal session id(%s) in client cookie", internalSession.ID)
				clientSession.Values["internal-session-id"] = internalSession.ID
				if cookieErr := clientSession.Save(httpReq, httpRes); cookieErr != nil {
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
				//render output
				httpRes.Header().Set("Content-Type", "text/html")
				if err = tmpl.ExecuteTemplate(httpRes, "page", data); err != nil {
					err = errors.Wrapf(err, "failed to exec template")
					return
				}
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

		log.Debugf("CLIENT cookie:")
		for key, val := range clientSession.Values {
			log.Debugf("  (%T)%+v : (%T)%+v", key, key, val, val)
		}

		//use new session if cannot load session
		internalSession = forms.Session{
			Data: map[string]interface{}{},
		}

		//get internal session identified by client session
		if internalSessionID, ok := clientSession.Values["internal-session-id"].(string); ok || internalSessionID != "" {
			//load existing session
			// log.Debugf("internal-session-id: \"%s\"", internalSessionID)
			res, err := msClient.Sync(
				ctx,
				ms.Address{
					Domain:    formsDomain,
					Operation: "get_session",
				},
				time.Millisecond*time.Duration(formsTTL),
				formsinterface.GetSessionRequest{
					ID: internalSessionID,
				},
				formsinterface.GetSessionResponse{})
			if err != nil {
				// log.Debugf("session.id(%s) not found", internalSessionID)
			} else {
				internalSession = res.(formsinterface.GetSessionResponse).Session
			}
		} //if has internal-session-id

		if internalSession.Data == nil {
			internalSession.Data = map[string]interface{}{}
		}

		log.Debugf("INTERNAL Session(id:%s)", internalSession.ID)
		for n, v := range internalSession.Data {
			log.Debugf("  %s: (%T)%v", n, v, v)
		}

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

		params := map[string]string{}
		for n, v := range httpReq.URL.Query() {
			params[n] = fmt.Sprintf("%v", v)
		}
		vars := mux.Vars(httpReq)
		for n, v := range vars {
			params[n] = v
		}

		switch httpReq.Method {
		case http.MethodGet:
			tmpl, data, err = getHdlr(ctx, &internalSession, params)
			if err != nil {
				err = errors.Wrapf(err, "get handler failed")
				return
			}

		case http.MethodPost:
			if err = httpReq.ParseForm(); err != nil {
				err = errors.Wrapf(err, "failed to parse the form data")
				return
			}
			log.Debugf("form data: %+v", httpReq.PostForm)

			tmpl, data, err = postHdlr(ctx, &internalSession, params, httpReq.PostForm)
			if err != nil {
				err = errors.Wrapf(err, "get handler failed")
				return
			}

		default:
			err = errors.Errorf("method not supported")
		}
	} //handlerFunc()
} //page2()

func updateInternalSession(ctx context.Context, session *forms.Session) error {
	if session.ID == "" {
		res, err := msClient.Sync(
			ctx,
			ms.Address{
				Domain:    formsDomain,
				Operation: "add_session",
			},
			time.Millisecond*time.Duration(formsTTL),
			formsinterface.AddSessionRequest{
				Session: *session,
			},
			formsinterface.AddSessionResponse{})
		if err != nil {
			log.Errorf("session add failed: %+v", err)
		} else {
			session.ID = res.(formsinterface.AddSessionResponse).Session.ID
			log.Debugf("Stored new session: %+v", res.(formsinterface.AddSessionResponse).Session)
		}
	} else {
		_, err := msClient.Sync(
			ctx,
			ms.Address{
				Domain:    formsDomain,
				Operation: "upd_session",
			},
			time.Millisecond*time.Duration(formsTTL),
			formsinterface.UpdSessionRequest{
				Session: *session,
			},
			nil)
		if err != nil {
			log.Errorf("failed to upd session.id(%s): %+v", session.ID, err)
		} else {
			log.Debugf("Updated session")
		}
	}
	return nil
} //updateInternalSession()
