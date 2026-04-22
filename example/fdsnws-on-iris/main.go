package main

import (
	"log"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/example/provider"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
)

func main() {
	service, err := fdsnws.NewFDSNWS(provider.New())
	if err != nil {
		log.Fatal(err)
	}

	app := iris.New()
	app.Use(logger.New())
	app.Any("/fdsnws/{path:path}", iris.FromStd(service))

	if err := app.Listen(":8080"); err != nil {
		log.Fatal(err)
	}
}
