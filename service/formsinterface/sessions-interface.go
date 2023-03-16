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
	DeviceID string `json:"device_id"`
}

func (req GetSessionRequest) Validate() error {
	if req.DeviceID == "" {
		return errors.Errorf("missing device-id")
	}
	return nil
}

type GetSessionResponse struct {
	Session forms.Session `json:"session"`
}

type UpdSessionRequest struct {
	DeviceID string        `json:"device_id"`
	Session  forms.Session `json:"session"`
}

func (req UpdSessionRequest) Validate() error {
	if req.DeviceID == "" {
		return errors.Errorf("missing device_id")
	}
	if err := req.Session.Validate(); err != nil {
		return errors.Wrapf(err, "invalid session")
	}
	return nil
}

type UpdSessionResponse struct {
	Session forms.Session `json:"session"`
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
