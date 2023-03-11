package forms

import (
	"time"

	"github.com/go-msvc/errors"
)

type Campaign struct {
	ID         string     `json:"id"`
	UserID     string     `json:"user_id"`
	CreateTime time.Time  `json:"create_time"`
	UpdateTime time.Time  `json:"update_time"`
	FormID     string     `json:"form_id" doc:"ID of form to be submitted"`
	StartTime  *time.Time `json:"start_time" doc:"Optional prevents submission before this time"`
	EndTime    *time.Time `json:"end_time" doc:"Optional prevents submission after this time"`
}

func (c Campaign) Validate() error {
	if c.UserID == "" {
		return errors.Errorf("missing user_id")
	}
	if c.FormID == "" {
		return errors.Errorf("missing form_id")
	}
	if c.StartTime != nil && c.EndTime != nil && c.StartTime.After(*c.EndTime) {
		return errors.Errorf("start_time:\"%s\" is after end_time:\"%s\"", *c.StartTime, *c.EndTime)
	}
	return nil
}
