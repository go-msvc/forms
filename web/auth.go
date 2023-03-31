package main

import (
	"context"
	"fmt"
	"html/template"
	"math/rand"
	"net/url"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/humans"
)

// loginEmailHandler is called when login-email-form is posted to send the OTP
func loginEmailHandler(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
	formData url.Values,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
) {
	log.Debugf("session=%v", session) //=nil?
	emailStr := formData.Get("email")
	var emailValue humans.Email
	if err := emailValue.Parse(emailStr); err != nil {
		//todo: show form again with error message...
		return nil, nil, errors.Errorf("invalid email \"%s\": %+v", emailStr, err)
	}
	//generate an OTP and store it internally with an expiry
	otp := newOtp()
	otpExpiry := time.Now().Add(time.Minute * 10)

	log.Errorf("NOT SENDING EMAIL")
	//send email with OTP
	// if err := email.Send(
	// 	email.Message{
	// 		From:        email.Email{Addr: "jan.semmelink@gmail.com", Name: "Forms"},
	// 		To:          []email.Email{{Addr: emailValue.String()}},
	// 		Cc:          []email.Email{},
	// 		Bcc:         []email.Email{},
	// 		Subject:     "Login OTP",
	// 		ContentType: "text/html",
	// 		Content: `<h1>Forms Login</h1>
	// 		<p>Enter the following OTP to login to forms.</p>
	// 		<h2>` + otp + `</h2>`,
	// 		AttachmentFilenames: []string{},
	// 	},
	// ); err != nil {
	// 	log.Errorf("failed to send email to \"%s\"", emailStr)
	// 	//return nil, nil, errors.Wrapf(err, "failed to send email to \"%s\"", emailStr)
	// }

	session.Authenticated = false
	session.Email = emailValue.String()
	session.Data["otp"] = otp
	session.Data["otp_expiry"] = otpExpiry.Format("2006-01-02T15:04:05Z")

	deviceID := ctx.Value(CtxDeviceID{}).(string)
	log.Debugf("device(%s).session(%s) sent email to %s with OTP:\"%s\"", deviceID, session.ID, emailStr, otp)

	//show OTP form
	otpFormData := map[string]interface{}{
		"Email":     emailStr,
		"OtpExpiry": otpExpiry.Format("2006-01-02 15:04:05"),
	}
	return loginOtpTemplate, otpFormData, nil
} //loginEmailHandler

// loginOtpHandler is called when login-otp-form is posted to verify OTP then redirect to user's home
func loginOtpHandler(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
	formData url.Values,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
) {
	log.Debugf("OTP Handler...")
	deviceID := ctx.Value(CtxDeviceID{}).(string)
	if session.Email == "" {
		log.Errorf("Missing session email: %+v", session)
		return loginEmailTemplate, map[string]interface{}{"error": "Missing email in session data"}, nil
	}
	otpExpiry, err := time.Parse("2006-01-02T15:04:05Z", session.Data["otp_expiry"].(string))
	if err != nil || otpExpiry.Before(time.Now()) {
		log.Errorf("OTP Expired:%v err:%v", otpExpiry, err)
		return loginEmailTemplate, map[string]interface{}{"error": "OTP expired"}, nil
	}
	expectedOtp := session.Data["otp"].(string)
	if expectedOtp == "" {
		log.Errorf("OTP Expired:%v err:%v", otpExpiry, err)
		log.Errorf("OTP=\"%s\"", expectedOtp)
		return loginEmailTemplate, map[string]interface{}{"error": "OTP not yet sent."}, nil
	}

	//ready for OTP check
	enteredOtp := formData.Get("otp")
	log.Errorf("entered otp=\"%s\"", enteredOtp)
	if enteredOtp == "" {
		return loginOtpTemplate, map[string]interface{}{"error": "OTP not yet entered."}, nil
	}

	//todo: uncomment - commented out so can login with any OTP while testing
	// if enteredOtp != expectedOtp {
	// 	log.Errorf("entered otp=\"%s\" != expectedOtp=\"%s\"", enteredOtp, expectedOtp)
	// 	return loginOtpTemplate, map[string]interface{}{"error": "Wrong OTP. Please try again."}, nil
	// }
	log.Errorf("entered CORRECT otp=\"%s\"", enteredOtp)

	//correct OTP - user is now logged in
	session.Authenticated = true

	//associate the device id with the email address
	//todo: if user has another authenticated session, switch to that session...
	//let it be done in back-end, associate device with that session, and email with this session if no session yet, then return the final session to here

	//todo...
	log.Errorf("Not yet storing device-id:\"%s\" against email:\"%s\"", deviceID, "?") //emailValue.String())

	//update session values to clear the expected OTP
	session.Data["otp"] = ""
	session.Data["opt_expiry"] = nil

	//go to page originally requested or go to user's home
	log.Debugf("Logged in")
	if targetURL, ok := ctx.Value(CtxTargetURL{}).(string); ok && targetURL != "" {
		log.Debugf("Go to target-url: %s", targetURL)
		return nil, nil, ErrorRedirect(targetURL)
	}
	log.Debugf("Go to user's home")
	return userHomeTemplate, nil, nil
} //loginOtpHandler

func logoutHandler(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
) {
	//this logs out all devices! It is safest to asume that is what user wants
	session.Authenticated = false
	return homeTemplate, nil, nil
}

func newOtp() string {
	otp := ""
	for i := 0; i < 4; i++ {
		n := rand.Intn(26)
		otp += fmt.Sprintf("%c", 'A'+n)
	}
	return otp
}
