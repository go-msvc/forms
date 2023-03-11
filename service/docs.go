package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/google/uuid"
)

var docsDir string

func init() {
	docsDir = os.Getenv("DOCS_DIR")
	if docsDir == "" {
		docsDir = "./docs"
	}
	if err := os.MkdirAll(docsDir, 0770); err != nil && err != os.ErrExist {
		panic(fmt.Sprintf("Cannot access DOCS_DIR=%s: %+v", docsDir, err))
	}
}

func addDoc(ctx context.Context, req formsinterface.AddDocRequest) (*formsinterface.AddDocResponse, error) {
	if req.Doc.ID != "" {
		return nil, errors.Errorf("doc.id=%s may not be specified when adding a doc", req.Doc.ID)
	}
	if req.Doc.Rev != 0 {
		return nil, errors.Errorf("doc.rev=%d may not be specified when adding a doc", req.Doc.Rev)
	}
	req.Doc.ID = uuid.New().String()
	req.Doc.Rev = 1
	req.Doc.Timestamp = time.Now()

	if err := saveDoc(req.Doc); err != nil {
		return nil, errors.Wrapf(err, "failed to save doc")
	}
	return &formsinterface.AddDocResponse{
		Doc: req.Doc,
	}, nil
} //addDoc()

func getDoc(ctx context.Context, req formsinterface.GetDocRequest) (*formsinterface.GetDocResponse, error) {
	if req.ID == "" {
		return nil, errors.Errorf("id must be specified when getting a doc")
	}
	existingDoc, err := loadDoc(req.ID, req.Rev)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load existing doc")
	}
	return &formsinterface.GetDocResponse{
		Doc: existingDoc,
	}, nil
} //getDoc()

func updDoc(ctx context.Context, req formsinterface.UpdDocRequest) (*formsinterface.UpdDocResponse, error) {
	if req.Doc.ID == "" {
		return nil, errors.Errorf("doc.id must be specified when updating a doc")
	}
	if req.Doc.Rev != 0 {
		return nil, errors.Errorf("doc.rev=%d may not be specified when updating a doc", req.Doc.Rev)
	}

	existingDoc, err := loadDoc(req.Doc.ID, 0) //0 for latest doc
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load existing doc")
	}
	req.Doc.Rev = existingDoc.Rev + 1
	req.Doc.Timestamp = time.Now()
	if err := saveDoc(req.Doc); err != nil {
		return nil, errors.Wrapf(err, "failed to save doc")
	}
	return &formsinterface.UpdDocResponse{
		Doc: req.Doc,
	}, nil
} //updDoc()

func delDoc(ctx context.Context, req formsinterface.DelDocRequest) (*formsinterface.DelDocResponse, error) {
	docDir := docsDir + "/" + req.ID
	if err := os.RemoveAll(docDir); err != nil {
		return nil, errors.Wrapf(err, "failed to remove doc")
	}
	return &formsinterface.DelDocResponse{}, nil
}

func findDoc(ctx context.Context, req formsinterface.FindDocRequest) (*formsinterface.FindDocResponse, error) {
	//should only see docs that you own or shared with you...
	return nil, MyError{Message: "NYI"}
}

func saveDoc(f forms.Doc) error {
	docDir := docsDir + "/" + f.ID
	if err := os.MkdirAll(docDir, 0770); err != nil && err != os.ErrExist {
		return errors.Wrapf(err, "cannot make doc dir %s", docDir)
	}

	filename := fmt.Sprintf("%s/latest.json", docDir)
	latestFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filename)
	}
	defer latestFile.Close()
	if err := json.NewEncoder(latestFile).Encode(f); err != nil {
		return errors.Wrapf(err, "failed to save latest doc")
	}

	filename = fmt.Sprintf("%s/rev_%d.json", docDir, f.Rev)
	revFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filename)
	}
	defer revFile.Close()
	if err := json.NewEncoder(revFile).Encode(f); err != nil {
		return errors.Wrapf(err, "failed to save rev doc")
	}
	return nil
} //saveDoc()

func loadDoc(id string, rev int) (forms.Doc, error) {
	docDir := docsDir + "/" + id
	var filename string
	if rev == 0 {
		filename = fmt.Sprintf("%s/latest.json", docDir)
	} else {
		filename = fmt.Sprintf("%s/rev_%d.json", docDir, rev)
	}

	docFile, err := os.Open(filename)
	if err != nil {
		return forms.Doc{}, errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer docFile.Close()
	var f forms.Doc
	if err := json.NewDecoder(docFile).Decode(&f); err != nil {
		return forms.Doc{}, errors.Wrapf(err, "failed to load latest doc")
	}
	return f, nil
} //loadDoc()
