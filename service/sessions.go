package main

import (
	"context"
	"sync"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/go-msvc/logger"
	"github.com/google/uuid"
)

//todo: cleanup expired sessions
//todo: cleanup expired devices and ask new login when used again...
//todo: store centrally - like everything else :-) currently in memory

var (
	log            = logger.New().WithLevel(logger.LevelDebug)
	sessionsMutex  sync.Mutex
	deviceByID     = map[string]*forms.Device{}
	sessionByEmail = map[string]*forms.Session{}
	sessionByID    = map[string]*forms.Session{}
)

func addSession(ctx context.Context, req formsinterface.AddSessionRequest) (*formsinterface.AddSessionResponse, error) {
	if req.Session.ID != "" {
		return nil, errors.Errorf("session.id=%s may not be specified when adding a session", req.Session.ID)
	}
	req.Session.ID = uuid.New().String()
	req.Session.TimeCreated = time.Now()
	req.Session.TimeUpdated = time.Now()
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	sessionByID[req.Session.ID] = &req.Session
	return &formsinterface.AddSessionResponse{
		Session: req.Session,
	}, nil
} //addSession()

func getSession(ctx context.Context, req formsinterface.GetSessionRequest) (*formsinterface.GetSessionResponse, error) {
	//device is created if not found
	device, ok := deviceByID[req.DeviceID]
	if !ok {
		device = &forms.Device{
			ID:          req.DeviceID,
			TimeCreated: time.Now(),
			Name:        "My Device",
			SessionID:   "",
		}
		deviceByID[req.DeviceID] = device
		log.Debugf("device(%s) is NEW", device.ID)
	}
	device.TimeLast = time.Now()

	var session *forms.Session
	if device.SessionID != "" {
		session, ok = sessionByID[device.SessionID]
		if !ok {
			device.SessionID = ""
		}
	}
	if device.SessionID == "" {
		//create a new unauthenticated blank session
		session = &forms.Session{
			ID:            uuid.New().String(),
			Authenticated: false,
			Email:         "",
			TimeCreated:   time.Now(),
			TimeUpdated:   time.Now(),
			Data:          map[string]interface{}{},
		}
		sessionByID[session.ID] = session
		device.SessionID = session.ID
		log.Debugf("device(%s).session(%s) created", device.ID, device.SessionID)
	} else {
		log.Debugf("device(%s).session(%s) existed", device.ID, device.SessionID)
	}
	return &formsinterface.GetSessionResponse{
		Session: *session,
	}, nil
} //getSession()

func updSession(ctx context.Context, req formsinterface.UpdSessionRequest) (*formsinterface.UpdSessionResponse, error) {
	if req.DeviceID == "" {
		return nil, errors.Errorf("device_id must be specified when updating a session")
	}
	if req.Session.ID == "" {
		return nil, errors.Errorf("session.id must be specified when updating a session")
	}

	device, ok := deviceByID[req.DeviceID]
	if !ok {
		return nil, errors.Errorf("device.id not found")
	}
	if device.SessionID != req.Session.ID {
		return nil, errors.Errorf("device.id(%s) cannot update session.id(%s)", req.DeviceID, req.Session.ID)
	}
	device.TimeLast = time.Now()

	existingSession, ok := sessionByID[req.Session.ID]
	if !ok {
		return nil, errors.Errorf("session not found")
	}
	if req.Session.Email != "" {
		existingSession.Email = req.Session.Email
	}

	if req.Session.Email != "" && req.Session.Authenticated && !existingSession.Authenticated {
		//logged in with a temp session
		//if already has authenticated session for this email, switch over to that session
		//else make this the authenticated session
		if authenticatedSession, ok := sessionByEmail[req.Session.Email]; ok && authenticatedSession.Authenticated && authenticatedSession.Email == req.Session.Email {
			//switch and delete the temp session
			delete(sessionByID, req.Session.ID)
			existingSession = authenticatedSession
			device.SessionID = authenticatedSession.ID
			log.Debugf("device(%s) authenticated and joins session(%s) for email(%s)", device.ID, existingSession.ID, existingSession.Email)
		} else {
			existingSession.Authenticated = true
			existingSession.Email = req.Session.Email
			sessionByEmail[req.Session.Email] = existingSession
			log.Debugf("device(%s) authenticated and new session(%s) for email(%s)", device.ID, existingSession.ID, existingSession.Email)
		}
		existingSession.TimeUpdated = time.Now()
	}

	//update session data
	for n, v := range req.Session.Data {
		existingSession.Data[n] = v
	}
	existingSession.TimeUpdated = time.Now()

	return &formsinterface.UpdSessionResponse{
		Session: *existingSession,
	}, nil
} //updSession()

func delSession(ctx context.Context, req formsinterface.DelSessionRequest) error {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	delete(sessionByID, req.ID)
	return nil
} //delSession()
