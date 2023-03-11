package forms

import (
	"time"

	"github.com/go-msvc/errors"
)

type Doc struct {
	ID        string                 `json:"id,omitempty"`
	Rev       int                    `json:"rev,omitempty"`
	Timestamp time.Time              `json:"timestamp" doc:"Time when the doc revision was created"`
	FormID    string                 `json:"form_id"`
	FormRev   int                    `json:"form_rev"`
	Data      map[string]interface{} `json:"data,omitempty" doc:"Submitted form data. Keys defined as name fields in the form."`
}

func (f *Doc) Validate() error {
	if f.Rev < 0 {
		return errors.Errorf("negative rev:%d", f.Rev)
	}
	if f.FormID == "" {
		return errors.Errorf("missing form_id")
	}
	if f.FormRev < 1 {
		return errors.Errorf("invalid form_rev:%d", f.FormRev)
	}
	return nil
} //Doc.Validate()

//todo: maintain foreign key between doc and form - but only one moved to a database...
