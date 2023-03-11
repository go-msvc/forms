package formsinterface

import (
	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
)

type AddCampaignRequest struct {
	Campaign forms.Campaign `json:"campaign"`
}

func (req AddCampaignRequest) Validate() error {
	if err := req.Campaign.Validate(); err != nil {
		return errors.Wrapf(err, "invalid campaign")
	}
	return nil
}

type AddCampaignResponse struct {
	Campaign forms.Campaign `json:"campaign"`
}

type GetCampaignRequest struct {
	ID string `json:"id"`
}

func (req GetCampaignRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type GetCampaignResponse struct {
	Campaign forms.Campaign `json:"campaign"`
}

type UpdCampaignRequest struct {
	Campaign forms.Campaign `json:"campaign"`
}

func (req UpdCampaignRequest) Validate() error {
	if err := req.Campaign.Validate(); err != nil {
		return errors.Wrapf(err, "invalid campaign")
	}
	return nil
}

type UpdCampaignResponse struct {
	Campaign forms.Campaign `json:"campaign"`
}

type DelCampaignRequest struct {
	ID string `json:"id"`
}

func (req DelCampaignRequest) Validate() error {
	if req.ID == "" {
		return errors.Errorf("missing id")
	}
	return nil
}

type DelCampaignResponse struct{}

type FindCampaignRequest struct {
	UserID string `json:"user_id"`
}

type FindCampaignResponse struct {
	//todo: list of campaigns
}

type CampaignNotification struct {
	CampaingID string `json:"campaign_id"`
	DocID      string `json:"doc_id"`
}
