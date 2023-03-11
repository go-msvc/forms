package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/google/uuid"
)

var formsDir string

func init() {
	formsDir = os.Getenv("FORMS_DIR")
	if formsDir == "" {
		formsDir = "./forms"
	}
	if err := os.MkdirAll(formsDir, 0770); err != nil && err != os.ErrExist {
		panic(fmt.Sprintf("Cannot access FORMS_DIR=%s: %+v", formsDir, err))
	}
}

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

func addForm(ctx context.Context, req AddFormRequest) (*AddFormResponse, error) {
	if req.Form.ID != "" {
		return nil, errors.Errorf("form.id=%s may not be specified when adding a form", req.Form.ID)
	}
	if req.Form.Rev != 0 {
		return nil, errors.Errorf("form.rev=%d may not be specified when adding a form", req.Form.Rev)
	}
	req.Form.ID = uuid.New().String()
	req.Form.Rev = 1

	if err := saveForm(req.Form); err != nil {
		return nil, errors.Wrapf(err, "failed to save form")
	}
	return &AddFormResponse{
		Form: req.Form,
	}, nil
} //addForm()

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

func getForm(ctx context.Context, req GetFormRequest) (*GetFormResponse, error) {
	if req.ID == "" {
		return nil, errors.Errorf("id must be specified when getting a form")
	}
	existingForm, err := loadForm(req.ID, req.Rev)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load existing form")
	}
	return &GetFormResponse{
		Form: existingForm,
	}, nil
} //getForm()

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

func updForm(ctx context.Context, req UpdFormRequest) (*UpdFormResponse, error) {
	if req.Form.ID == "" {
		return nil, errors.Errorf("form.id must be specified when updating a form")
	}
	if req.Form.Rev != 0 {
		return nil, errors.Errorf("form.rev=%d may not be specified when updating a form", req.Form.Rev)
	}

	existingForm, err := loadForm(req.Form.ID, 0) //0 for latest form
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load existing form")
	}
	req.Form.Rev = existingForm.Rev + 1
	if err := saveForm(req.Form); err != nil {
		return nil, errors.Wrapf(err, "failed to save form")
	}
	return &UpdFormResponse{
		Form: req.Form,
	}, nil
} //updForm()

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

func delForm(ctx context.Context, req DelFormRequest) (*DelFormResponse, error) {
	formDir := formsDir + "/" + req.ID
	if err := os.RemoveAll(formDir); err != nil {
		return nil, errors.Wrapf(err, "failed to remove form")
	}
	return &DelFormResponse{}, nil
}

type FindFormRequest struct{}

type FindFormResponse struct{}

func findForm(ctx context.Context, req AddFormRequest) (*AddFormResponse, error) {
	//should only see forms that you own or shared with you...
	return nil, MyError{Message: "NYI"}
}

func saveForm(f forms.Form) error {
	formDir := formsDir + "/" + f.ID
	if err := os.Mkdir(formDir, 0770); err != nil && err != os.ErrExist {
		return errors.Wrapf(err, "cannot make form dir %s", formDir)
	}

	filename := fmt.Sprintf("%s/latest.json", formDir)
	latestFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filename)
	}
	defer latestFile.Close()
	if err := json.NewEncoder(latestFile).Encode(f); err != nil {
		return errors.Wrapf(err, "failed to save latest form")
	}

	filename = fmt.Sprintf("%s/rev_%d.json", formDir, f.Rev)
	revFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filename)
	}
	defer revFile.Close()
	if err := json.NewEncoder(revFile).Encode(f); err != nil {
		return errors.Wrapf(err, "failed to save rev form")
	}
	return nil
} //saveForm()

func loadForm(id string, rev int) (forms.Form, error) {
	formDir := formsDir + "/" + id
	var filename string
	if rev == 0 {
		filename = fmt.Sprintf("%s/latest.json", formDir)
	} else {
		filename = fmt.Sprintf("%s/rev_%d.json", formDir, rev)
	}

	formFile, err := os.Open(filename)
	if err != nil {
		return forms.Form{}, errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer formFile.Close()
	var f forms.Form
	if err := json.NewDecoder(formFile).Decode(&f); err != nil {
		return forms.Form{}, errors.Wrapf(err, "failed to load latest form")
	}
	return f, nil
} //loadForm()
