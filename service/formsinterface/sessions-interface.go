package formsinterface

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
)

type AddSessionRequest struct {
	Session forms.Session `json:"session"`
}

func (req AddSessionRequest) Validate() error {
	if err := req.Session.Validate(); err != nil {
		return errors.Wrapf(err, "invalid session")
	}
	return nil
}

type AddSessionResponse struct {
	Session forms.Session `json:"session"`
}

type GetSessionRequest struct {
	ID  string `json:"id"`
	Rev int    `json:"rev" session:"Use 0 for the latest version of the session"`
}

func (req GetSessionRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type GetSessionResponse struct {
	Session forms.Session `json:"session"`
}

type UpdSessionRequest struct {
	Session forms.Session `json:"session"`
}

func (req UpdSessionRequest) Validate() error {
	if err := req.Session.Validate(); err != nil {
		return errors.Wrapf(err, "invalid session")
	}
	return nil
}

type DelSessionRequest struct {
	ID string `json:"id"`
}

func (req DelSessionRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type DelSessionResponse struct{}

type FindSessionRequest struct{}

type FindSessionResponse struct{}
