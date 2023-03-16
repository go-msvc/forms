package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"net/url"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/go-msvc/utils/ms"
)

func showCampaign(ctx context.Context, session *forms.Session, params map[string]string) (*template.Template, interface{}, error) {
	log.Debugf("showCampaign(%+v)", params)
	campaign, form, err := loadCampaign(ctx, params["id"])
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to load campaign")
	}

	//set data in session that will be needed when the form is submitted
	session.Data["campaign_id"] = campaign.ID
	session.Data["form_id"] = form.ID
	session.Data["form_rev"] = form.Rev

	//render markdown in the form to HTML
	form.Header = renderHeaderHTML(form.Header)
	for i, s := range form.Sections {
		s.Header = renderHeaderHTML(s.Header)
		for itemIndex, item := range s.Items {
			if item.Header != nil {
				*item.Header = renderHeaderHTML(*item.Header)
			}
			if item.Field != nil {
				item.Field.Header = renderHeaderHTML(item.Field.Header)
			}
			if item.Image != nil {
				item.Image.Header = renderHeaderHTML(item.Image.Header)
			}
			if item.Table != nil {
				item.Table.Header = renderHeaderHTML(item.Table.Header)
			}
			if item.Sub != nil {
				item.Sub.Header = renderHeaderHTML(item.Sub.Header)
			}
			s.Items[itemIndex] = item
		} //for each item
		form.Sections[i] = s
	} //for each section

	//set values needed in the form
	form.Action = fmt.Sprintf("/campaign/%s", campaign.ID)
	form.CampaignID = campaign.ID

	//load form template (todo: use global already loaded template when not in dev)
	formTemplate := loadTemplates([]string{"form", "page"})
	return formTemplate, form, nil
} //showCampaign()

func postCampaign(ctx context.Context, session *forms.Session, params map[string]string, formData url.Values) (*template.Template, interface{}, error) {
	log.Debugf("postCampaign(%+v)", params)

	id := session.Data["campaign_id"].(string)
	campaign, _, err := loadCampaign(ctx, id)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to load campaign")
	}

	doc, err := postForm(ctx, session, formData)
	if err != nil {
		log.Errorf("failed to post submitted form: %+v", err)
		return nil, nil, errors.Wrapf(err, "failed to submit the form data")
	}
	log.Debugf("Submitted: %+v", doc)

	//send campaign notification
	notification := formsinterface.CampaignNotification{
		CampaingID: campaign.ID, //todo: should come from ctx session data
		DocID:      doc.ID,
	}
	jsonNotification, _ := json.Marshal(notification)
	if _, err := redisClient.LPush(ctx, campaign.ID, jsonNotification).Result(); err != nil {
		return nil, nil, errors.Wrapf(err, "failed to send for processing")
	}

	//show details of submitted documents
	return campaignSubmittedTemplate, map[string]interface{}{
		"CampaignID": campaign.ID,
	}, nil
} //postCampaign()

func loadCampaign(ctx context.Context, id string) (forms.Campaign, forms.Form, error) {
	res, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_campaign",
		},
		formsTTL,
		formsinterface.GetCampaignRequest{
			ID: id,
		},
		formsinterface.GetCampaignResponse{})
	if err != nil {
		return forms.Campaign{}, forms.Form{}, errors.Wrapf(err, "campaign.id(%s) not found", id)
	}
	campaign := res.(formsinterface.GetCampaignResponse).Campaign

	res, err = msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_form",
		},
		formsTTL,
		formsinterface.GetFormRequest{
			ID: campaign.FormID,
		},
		formsinterface.GetFormResponse{})
	if err != nil {
		return forms.Campaign{}, forms.Form{}, errors.Wrapf(err, "campaign.form.id(%s) not found", id)
	}
	form := res.(formsinterface.GetFormResponse).Form

	if campaign.StartTime.After(time.Now()) {
		return forms.Campaign{}, forms.Form{}, errors.Wrapf(err, "campaign(%s) only starts at %v", campaign.ID, campaign.StartTime)
	}
	if campaign.EndTime.Before(time.Now()) {
		return forms.Campaign{}, forms.Form{}, errors.Wrapf(err, "campaign(%s) ended at %v", campaign.ID, campaign.EndTime)
	}
	return campaign, form, nil
} //loadCampaign()
