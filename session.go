package forms

import (
	"time"
)

// internal session data stored in the back-end
type Session struct {
	ID            string                 `json:"id" doc:"Assigned when session is created for this email"`
	Authenticated bool                   `json:"bool" doc:"Set true after authenticated"`
	Email         string                 `json:"email" doc:"Email is unique for each session. Only one session per email and created after entered OTP on a device."`
	TimeCreated   time.Time              `json:"time_created"`
	TimeUpdated   time.Time              `json:"time_updated"`
	Data          map[string]interface{} `json:"data"`
}

func (s Session) Validate() error {
	return nil
} //Session.Validate()

type Device struct {
	ID          string    `json:"id" doc:"UUID generated on first use of device or after browser history was cleared. Stored in browser cookie to identify the device."`
	TimeCreated time.Time `json:"time_created" doc:"Time when the device ID was generated."`
	TimeLast    time.Time `json:"time_last" doc:"Time when it was last used to attempt to retrieve a session"`
	Name        string    `json:"name,omitempty" doc:"User's way to identify the device. User can enter a name like \"MyPhone\""`
	SessionID   string    `json:"session-id,omitempty" doc:"Session associated with this device"`
}
