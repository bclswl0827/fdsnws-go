package main

import (
	"log"
	"net/http"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/example/provider"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	service, err := fdsnws.NewFDSNWS(provider.New())
	if err != nil {
		log.Fatal(err)
	}

	app := chi.NewRouter()

	app.Use(middleware.RequestID)
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)

	app.Mount("/fdsnws", service)

	listen := ":8080"
	log.Printf("http server started at %s", listen)
	log.Fatal(http.ListenAndServe(listen, app))
}
