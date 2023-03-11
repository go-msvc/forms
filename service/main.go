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
	)
	if err := config.Load(); err != nil {
		panic(err)
	}
	ms.Configure()
	ms.Serve()
}
