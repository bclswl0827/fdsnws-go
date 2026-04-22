package main

import (
	"log"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/example/provider"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	service, err := fdsnws.NewFDSNWS(provider.New())
	if err != nil {
		log.Fatal(err)
	}
	e := echo.New()
	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())
	e.Any("/fdsnws/*", echo.WrapHandler(service))
	e.Logger.Fatal(e.Start(":8080"))
}
