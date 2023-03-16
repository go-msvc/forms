package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/go-msvc/errors"
	"github.com/go-msvc/forms"
	"github.com/go-msvc/forms/service/formsinterface"
	"github.com/google/uuid"
)

var campaignsDir string

func init() {
	campaignsDir = os.Getenv("CAMPAIGNS_DIR")
	if campaignsDir == "" {
		campaignsDir = "./campaigns"
	}
	if err := os.MkdirAll(campaignsDir, 0770); err != nil && err != os.ErrExist {
		panic(fmt.Sprintf("Cannot access CAMPAIGNS_DIR=%s: %+v", campaignsDir, err))
	}
}

func addCampaign(ctx context.Context, req formsinterface.AddCampaignRequest) (*formsinterface.AddCampaignResponse, error) {
	if req.Campaign.ID != "" {
		return nil, errors.Errorf("campaign.id=%s may not be specified when adding a campaign", req.Campaign.ID)
	}
	req.Campaign.ID = uuid.New().String()
	req.Campaign.CreateTime = time.Now()
	req.Campaign.UpdateTime = time.Now()
	if err := saveCampaign(req.Campaign); err != nil {
		return nil, errors.Wrapf(err, "failed to save campaign")
	}
	return &formsinterface.AddCampaignResponse{
		Campaign: req.Campaign,
	}, nil
} //addCampaign()

func getCampaign(ctx context.Context, req formsinterface.GetCampaignRequest) (*formsinterface.GetCampaignResponse, error) {
	if req.ID == "" {
		return nil, errors.Errorf("id must be specified when getting a campaign")
	}
	existingCampaign, err := loadCampaign(req.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load existing campaign")
	}
	return &formsinterface.GetCampaignResponse{
		Campaign: existingCampaign,
	}, nil
} //getCampaign()

func updCampaign(ctx context.Context, req formsinterface.UpdCampaignRequest) (*formsinterface.UpdCampaignResponse, error) {
	if req.Campaign.ID == "" {
		return nil, errors.Errorf("campaign.id must be specified when updating a campaign")
	}
	_, err := loadCampaign(req.Campaign.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load existing campaign")
	}
	req.Campaign.UpdateTime = time.Now()
	if err := saveCampaign(req.Campaign); err != nil {
		return nil, errors.Wrapf(err, "failed to save campaign")
	}
	return &formsinterface.UpdCampaignResponse{
		Campaign: req.Campaign,
	}, nil
} //updCampaign()

func delCampaign(ctx context.Context, req formsinterface.DelCampaignRequest) (*formsinterface.DelCampaignResponse, error) {
	campaignDir := campaignsDir + "/" + req.ID
	if err := os.RemoveAll(campaignDir); err != nil {
		return nil, errors.Wrapf(err, "failed to remove campaign")
	}
	return &formsinterface.DelCampaignResponse{}, nil
}

func findCampaigns(ctx context.Context, req formsinterface.FindCampaignRequest) (*formsinterface.FindCampaignResponse, error) {
	//should only see campaigns that you own or shared with you...
	return nil, MyError{Message: "NYI"}
}

func saveCampaign(f forms.Campaign) error {
	campaignDir := campaignsDir + "/" + f.ID
	if err := os.MkdirAll(campaignDir, 0770); err != nil && err != os.ErrExist {
		return errors.Wrapf(err, "cannot make campaign dir %s", campaignDir)
	}
	filename := fmt.Sprintf("%s/latest.json", campaignDir)
	latestFile, err := os.Create(filename)
	if err != nil {
		return errors.Wrapf(err, "failed to create file %s", filename)
	}
	defer latestFile.Close()
	if err := json.NewEncoder(latestFile).Encode(f); err != nil {
		return errors.Wrapf(err, "failed to save latest campaign")
	}
	return nil
} //saveCampaign()

func loadCampaign(id string) (forms.Campaign, error) {
	campaignDir := campaignsDir + "/" + id
	var filename string
	filename = fmt.Sprintf("%s/latest.json", campaignDir)
	campaignFile, err := os.Open(filename)
	if err != nil {
		return forms.Campaign{}, errors.Wrapf(err, "failed to open file %s", filename)
	}
	defer campaignFile.Close()
	var f forms.Campaign
	if err := json.NewDecoder(campaignFile).Decode(&f); err != nil {
		return forms.Campaign{}, errors.Wrapf(err, "failed to load latest campaign")
	}
	return f, nil
} //loadCampaign()
