package main

import (
	"log"
	"net/http"
	"time"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/example/provider"
)

func main() {
	service, err := fdsnws.NewFDSNWS(provider.New())
	if err != nil {
		log.Fatal(err)
	}

	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", loggingMiddleware(service)))
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
