package formsinterface

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
)

type AddFormRequest struct {
	Form forms.Form `json:"form"`
}

func (req AddFormRequest) Validate() error {
	if err := req.Form.Validate(); err != nil {
		return errors.Wrapf(err, "invalid form")
	}
	return nil
}

type AddFormResponse struct {
	Form forms.Form `json:"form"`
}

type GetFormRequest struct {
	ID  string `json:"id"`
	Rev int    `json:"rev" doc:"Use 0 for the latest version of the form"`
}

func (req GetFormRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type GetFormResponse struct {
	Form forms.Form `json:"form"`
}

type UpdFormRequest struct {
	Form forms.Form `json:"form"`
}

func (req UpdFormRequest) Validate() error {
	if err := req.Form.Validate(); err != nil {
		return errors.Wrapf(err, "invalid form")
	}
	return nil
}

type UpdFormResponse struct {
	Form forms.Form `json:"form"`
}

type DelFormRequest struct {
	ID string `json:"id"`
}

func (req DelFormRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type DelFormResponse struct{}

type FindFormRequest struct{}

type FindFormResponse struct{}
