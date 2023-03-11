package main

import (
	"context"
	"sync"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/google/uuid"
)

var (
	sessionsMutex sync.Mutex
	sessionByID   = map[string]*forms.Session{}
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
	if req.ID == "" {
		return nil, errors.Errorf("id must be specified when getting a session")
	}
	existingSession, ok := sessionByID[req.ID]
	if !ok {
		return nil, errors.Errorf("session not found")
	}
	return &formsinterface.GetSessionResponse{
		Session: *existingSession,
	}, nil
} //getSession()

func updSession(ctx context.Context, req formsinterface.UpdSessionRequest) error {
	if req.Session.ID == "" {
		return errors.Errorf("session.id must be specified when updating a session")
	}
	existingSession, ok := sessionByID[req.Session.ID]
	if !ok {
		return errors.Errorf("session not found")
	}
	for n, v := range req.Session.Data {
		existingSession.Data[n] = v
	}
	existingSession.TimeUpdated = time.Now()
	return nil
} //updSession()

func delSession(ctx context.Context, req formsinterface.DelSessionRequest) error {
	sessionsMutex.Lock()
	defer sessionsMutex.Unlock()
	delete(sessionByID, req.ID)
	return nil
} //delSession()
