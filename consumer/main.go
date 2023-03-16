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
	"github.com/go-msvc/logger"
	"github.com/go-msvc/nats-utils"
	"github.com/go-msvc/utils/ms"
	"github.com/go-redis/redis/v8"
)

var (
	formsDomain = "forms"
	formsTTL    = time.Second * 1
	redisClient *redis.Client
	msClient    ms.Client
)

var log = logger.New().WithLevel(logger.LevelDebug)

func main() {
	//ms client
	msClientConfig := nats.ClientConfig{
		Config: nats.Config{
			Domain: "forms-web",
		},
	}
	if err := msClientConfig.Validate(); err != nil {
		panic(fmt.Sprintf("client config: %+v", err))
	}
	var err error
	msClient, err = msClientConfig.Create()
	if err != nil {
		panic(fmt.Sprintf("failed to create ms client: %+v", err))
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	key := os.Getenv("REDIS_CONSUMER_KEY")
	if key == "" {
		panic("REDIS_CONSUMER_KEY is not defined")
	}
	for {
		ctx := context.Background()
		result, err := redisClient.BRPop(ctx, 10*time.Second, key).Result()
		if err != nil {
			log.Debugf("Empty queue %s ...", key)
			continue
		}

		//result[0] = key that popped
		//result[1] = value popped
		var notification formsinterface.CampaignNotification
		if err := json.Unmarshal([]byte(result[1]), &notification); err != nil {
			log.Errorf("discard because cannot decode into Notification: %+v: %+v", string(result[1]), err)
		}
		go func(ctx context.Context, n formsinterface.CampaignNotification) {
			if err := process(ctx, result[0], n); err != nil {
				log.Errorf("failed to process: %+v: %+v", notification, err)
			}
		}(ctx, notification)
	} //for
} //main()

func process(ctx context.Context, key string, n formsinterface.CampaignNotification) error {
	log.Debugf("Processing: %+v: %+v", key, n)
	campaign, doc, err := loadCampaignDocument(ctx, n.CampaingID, n.DocID)
	if err != nil {
		return errors.Wrapf(err, "failed to load")
	}
	log.Debugf("campaign: %+v", campaign)
	log.Debugf("doc: %+v", doc)

	//todo: would be nice to call user logic here and cosume all documents
	//so not have to start a micro-service for each campaign
	//for this need to make notification key configurable, default to generic key...

	if campaign.Action == nil {

	}

	return errors.Errorf("NYI")
} //process()

func loadCampaignDocument(ctx context.Context, campaignID, docID string) (forms.Campaign, forms.Doc, error) {
	res, err := msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_campaign",
		},
		formsTTL,
		formsinterface.GetCampaignRequest{
			ID: campaignID,
		},
		formsinterface.GetCampaignResponse{})
	if err != nil {
		return forms.Campaign{}, forms.Doc{}, errors.Wrapf(err, "campaign.id(%s) not found", campaignID)
	}
	campaign := res.(formsinterface.GetCampaignResponse).Campaign

	res, err = msClient.Sync(
		ctx,
		ms.Address{
			Domain:    formsDomain,
			Operation: "get_doc",
		},
		time.Millisecond*time.Duration(formsTTL),
		formsinterface.GetDocRequest{
			ID: docID,
		},
		formsinterface.GetDocResponse{})
	if err != nil {
		return forms.Campaign{}, forms.Doc{}, errors.Wrapf(err, "doc.id(%s) not found", docID)
	}
	doc := res.(formsinterface.GetDocResponse).Doc
	return campaign, doc, nil
} //loadCampaignDocument()
