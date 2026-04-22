package provider

import (
	"embed"
	"fmt"

	fdsnws "github.com/bclswl0827/fdsnws-go"
)

//go:embed web
var builtinWeb embed.FS

func loadWebPage(path string) (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	data, err := builtinWeb.ReadFile(fmt.Sprintf("web%s", path))
	if err != nil {
		return fdsnws.PreparedResponse{
			StatusCode: 500,
			Body:       err.Error(),
			MIMEType:   fdsnws.MIMEText,
		}, fdsnws.ProviderErrorNone, err
	}

	return fdsnws.PreparedResponse{
		Body:       string(data),
		MIMEType:   "text/html",
		StatusCode: 200,
	}, fdsnws.ProviderErrorNone, nil
}

func (Provider) FDSNWSRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/index.html")
}

func (Provider) DataSelectRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/dataselect/index.html")
}

func (Provider) DataSelectVersionRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/dataselect/1/index.html")
}

func (Provider) EventRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/event/index.html")
}

func (Provider) EventVersionRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/event/1/index.html")
}

func (Provider) StationRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/station/index.html")
}

func (Provider) StationVersionRoot() (fdsnws.PreparedResponse, fdsnws.ProviderErrorKind, error) {
	return loadWebPage("/station/1/index.html")
}
