package forms

import (
	"time"

	"github.com/go-msvc/errors"
)

type Campaign struct {
	ID         string         `json:"id"`
	UserID     string         `json:"user_id"`
	CreateTime time.Time      `json:"create_time"`
	UpdateTime time.Time      `json:"update_time"`
	FormID     string         `json:"form_id" doc:"ID of form to be submitted"`
	StartTime  *time.Time     `json:"start_time" doc:"Optional prevents submission before this time"`
	EndTime    *time.Time     `json:"end_time" doc:"Optional prevents submission after this time"`
	Queue      string         `json:"queue" doc:"Queue where notification is sent. If not specified, default processing applied configured in action."`
	Action     CampaignAction `json:"action" doc:"What to do with submitted documents"`
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

type CampaignAction struct {
	Http *CampaignActionHttp `json:"http" doc:"Specify to call an HTTP end-point"`
	//Email ... send email message with summary in body and JSON attachments, or consider using a template and markdown to construct the message..."
	//MS ... call a micro-service to implement custom logic
	//Forward send to other REDIS queue(s)
}

type CampaignActionHttp struct {
	URL    string `json:"url" doc:"The URL may include {{.DocID}} and {{.CampaignID}} for substitution."`
	Method string `json:"method" doc:"When POST|PUT, body will be formsinterface.AddDocRequest with Content-Type:application/json"`
}
