package formsinterface

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
)

type AddDocRequest struct {
	Doc forms.Doc `json:"doc"`
}

func (req AddDocRequest) Validate() error {
	if err := req.Doc.Validate(); err != nil {
		return errors.Wrapf(err, "invalid doc")
	}
	return nil
}

type AddDocResponse struct {
	Doc forms.Doc `json:"doc"`
}

type GetDocRequest struct {
	ID  string `json:"id"`
	Rev int    `json:"rev" doc:"Use 0 for the latest version of the doc"`
}

func (req GetDocRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type GetDocResponse struct {
	Doc forms.Doc `json:"doc"`
}

type UpdDocRequest struct {
	Doc forms.Doc `json:"doc"`
}

func (req UpdDocRequest) Validate() error {
	if err := req.Doc.Validate(); err != nil {
		return errors.Wrapf(err, "invalid doc")
	}
	return nil
}

type UpdDocResponse struct {
	Doc forms.Doc `json:"doc"`
}

type DelDocRequest struct {
	ID string `json:"id"`
}

func (req DelDocRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type DelDocResponse struct{}

type FindDocRequest struct{}

type FindDocResponse struct{}
