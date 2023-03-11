package forms

import "time"

// internal session data stored in the back-end
type Session struct {
	ID          string                 `json:"id"`
	TimeCreated time.Time              `json:"time_created"`
	TimeUpdated time.Time              `json:"time_updated"`
	Data        map[string]interface{} `json:"data"`
}

func (s Session) Validate() error {
	return nil
}

//todo: cleanup expired sessions
//todo: store centrally - like everything else :-)
