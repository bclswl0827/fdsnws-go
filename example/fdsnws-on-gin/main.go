package main

import (
	"log"
	"net/http"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/example/provider"
	"github.com/gin-gonic/gin"
)

func main() {
	service, err := fdsnws.NewFDSNWS(provider.New())
	if err != nil {
		log.Fatal(err)
	}

	router := newRouter(service)
	log.Fatal(router.Run(":8080"))
}

func newRouter(handler http.Handler) *gin.Engine {
	router := gin.Default()
	router.Any("/fdsnws/*path", gin.WrapH(handler))
	return router
}
