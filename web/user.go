package main

import (
	"context"
	"html/template"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
)

func userHomeGetHandler(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
) {
	log.Debugf("Showing User's Home (params:%+v)", params)

	//todo: replace this test data
	pageData := UserHomeTmplData{
		Campaigns: []CampaignTmplData{
			{
				ID:                 "8df7997b-4a90-4784-a496-c0ec453c10f2",
				Title:              "C1",
				TimeCreated:        time.Now().Add(-time.Hour * 24),
				LastSubmissionTime: time.Now().Add(time.Hour),
				NrSubmissions:      10,
			},
			{
				ID:                 "8df7997b-4a90-4784-a496-c0ec453c10f2", //same as above
				Title:              "C2",
				TimeCreated:        time.Now().Add(-time.Hour * 224),
				LastSubmissionTime: time.Now().Add(time.Hour * 3),
				NrSubmissions:      20,
			},
		},
	}
	return userHomeTemplate, pageData, nil
} //userHomeGetHandler()

type UserHomeTmplData struct {
	Campaigns []CampaignTmplData
}

type CampaignTmplData struct {
	ID                 string
	Title              string
	TimeCreated        time.Time
	NrSubmissions      int
	LastSubmissionTime time.Time
}

func myCampaign(
	ctx context.Context,
	session *forms.Session,
	params map[string]string,
) (
	tmpl *template.Template,
	tmplData interface{},
	err error,
) {
	log.Debugf("Campaign Details (params:%+v)", params)

	c, f, err := loadCampaign(ctx, params["campaign_id"])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "campaign not loaded")
	}

	//todo: replace this test data
	pageData := CampaignTmplData{
		ID:                 c.ID,
		Title:              f.Title,
		TimeCreated:        c.CreateTime,
		LastSubmissionTime: time.Now().Add(time.Hour), //todo
		NrSubmissions:      10,                        //todo
	}
	return userCampaignTemplate, pageData, nil
} //myCampaign()
