package main

import (
	"log"
	"net/url"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/example/provider"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/fiber/v3/middleware/recover"
)

func main() {
	service, err := fdsnws.NewFDSNWS(provider.New())
	if err != nil {
		log.Fatal(err)
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(recover.New())

	app.All("/fdsnws/*", func(c fiber.Ctx) error {
		parsedURL, err := url.ParseRequestURI(c.OriginalURL())
		if err != nil {
			return err
		}
		bodyData := c.Body()
		body := make([]byte, len(bodyData))
		copy(body, bodyData)
		response := service.Handle(c.Context(), fdsnws.Request{
			Method:     c.Method(),
			Path:       c.Path(),
			RawQuery:   parsedURL.RawQuery,
			Query:      parsedURL.Query(),
			Headers:    c.GetReqHeaders(),
			Body:       body,
			RemoteAddr: c.IP(),
			URL:        c.OriginalURL(),
		})
		for key, value := range response.Headers {
			c.Set(key, value)
		}
		if response.MIMEType != "" {
			c.Set(fiber.HeaderContentType, response.MIMEType)
		}
		return c.Status(response.StatusCode).Send(response.Body)
	})

	log.Fatal(app.Listen(":8080"))
}
