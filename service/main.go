package main

import (
	"github.com/go-msvc/config"
	_ "github.com/go-msvc/nats-utils"
	"github.com/go-msvc/utils/ms"
)

func main() {
	config.AddSource("config.json", config.File("./config.json"))
	ms := ms.New(
		ms.WithOper("add_form", addForm),
		ms.WithOper("get_form", getForm),
		ms.WithOper("upd_form", updForm),
		ms.WithOper("del_form", delForm),
		ms.WithOper("find_forms", findForm),
		ms.WithOper("add_doc", addDoc),
		ms.WithOper("get_doc", getDoc),
		ms.WithOper("upd_doc", updDoc),
		ms.WithOper("del_doc", delDoc),
		ms.WithOper("find_docs", findDoc),
		ms.WithOper("add_campaign", addCampaign),
		ms.WithOper("get_campaign", getCampaign),
		ms.WithOper("upd_campaign", updCampaign),
		ms.WithOper("del_campaign", delCampaign),
		ms.WithOper("find_campaigns", findCampaigns),
	)
	if err := config.Load(); err != nil {
		panic(err)
	}
	ms.Configure()
	ms.Serve()
}
