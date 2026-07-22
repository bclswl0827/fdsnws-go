# fdsnws-go

FDSN Web Services protocol implementation in pure Go.

`fdsnws-go` provides framework-independent routing, request validation, FDSN error responses, WADL generation, and response encoding. Applications supply data through a typed Provider and may enable only the services they need.

## Features

- **DataSelect 1.1**: `query`, optional `queryauth`, `version`, and `application.wadl`; encodes Provider samples as miniSEED with [`mseedio`](https://github.com/bclswl0827/mseedio).
- **Station 1.1**: `query`, `version`, and `application.wadl`; supports StationXML and FDSN text output.
- **Event 1.2**: `query`, `catalogs`, `contributors`, `version`, and `application.wadl`; supports QuakeML and FDSN text output.
- **Modular services**: DataSelect, Station, and Event are independent `service.Service` components.
- **Framework independent**: use normalized `Request` and `Response` values or the optional `http.Handler` adapter.
- **Typed Providers**: Providers receive parsed queries, never HTTP or framework contexts.
- **Reusable XML packages**: `stationxml` and `quakeml` can parse, build, and encode documents outside the web service.
- **Provider clock**: protocol timestamps come from `Provider.Now()` and can be fixed in tests.

## Installation

Requires Go 1.25 or later.

```shell
$ go get github.com/bclswl0827/fdsnws-go
```

## Quick Usage

Embed `UnimplementedProvider` and implement only the methods needed by enabled modules:

```go
package main

import (
	"log"
	"net/http"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/stationxml"
)

type Provider struct {
	fdsnws.UnimplementedProvider
}

func (Provider) Station(query fdsnws.StationQuery) (*stationxml.Document, fdsnws.ProviderErrorKind, error) {
	document := stationxml.NewDocument()
	document.Source = "example"
	document.Networks = []stationxml.Network{{Code: "XX"}}
	return document, fdsnws.ProviderErrorNone, nil
}

func main() {
	server, err := fdsnws.NewFDSNWS(
		Provider{},
		fdsnws.WithModules(fdsnws.ModuleStation),
	)
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(http.ListenAndServe(":8080", server))
}
```

`WithModules` enables exactly the listed modules. Without it, all built-in modules are enabled.

## Provider Results

- `DataSelect` returns `DataSelectResult` records containing NSLC metadata, start time, sample rate, and `[]int32` samples.
- `Station` returns `*stationxml.Document` directly.
- `Event` returns `*quakeml.Document` directly.
- Service root methods return `PreparedResponse` with body, MIME type, and status code.
- `ProviderErrorClient` maps ordinary Provider errors to HTTP 400; `ProviderErrorServer` maps them to HTTP 500. Use `HTTPError` for an exact status and headers, or `ErrNoData` for an empty result.

The complete Provider also implements `Now() time.Time`. Embedding `UnimplementedProvider` supplies a UTC system clock; override `Now` to use an application clock.

## Authentication

`queryauth` is enabled only when an independent `Authenticator` is registered:

```go
server, err := fdsnws.NewFDSNWS(
	provider,
	fdsnws.WithAuthenticator(authenticator),
)
```

The authenticator receives method, request target, Authorization header, body, and remote address. On success, the resulting `Principal` is available as `DataSelectQuery.Principal`; credentials are never passed to the data Provider. Return an `HTTPError` with status 401 and `WWW-Authenticate` for a Digest challenge.

Without an authenticator, `/fdsnws/dataselect/1/queryauth` returns 404 and is omitted from WADL.

## StationXML and QuakeML

Both XML packages support byte slices and streams:

```go
stationDocument, err := stationxml.Parse(stationData)
stationData, err = stationxml.Marshal(stationDocument)

eventDocument, err := quakeml.Decode(reader)
err = quakeml.Encode(writer, eventDocument)
```

Use `stationxml.NewDocument()` or `quakeml.NewDocument(publicID)` to create documents with default namespaces.

## Examples

Runnable adapters are included for `net/http`, Gin, Echo, Fiber, Chi, and Iris. They share the Provider in `example/provider`.

```shell
$ go run ./example/fdsnws-on-gin
```

```shell
$ curl 'http://localhost:8080/fdsnws/dataselect/1/query?network=XX&station=DEMO&location=00&channel=BHZ&starttime=2024-01-01T00:00:00&endtime=2024-01-01T00:00:05' --output example.mseed
```

## Project Layout

- `service/` — common contracts and DataSelect, Station, and Event components.
- `stationxml/` — reusable FDSN StationXML model and codec.
- `quakeml/` — reusable QuakeML model and codec.
- `utils/` — shared request parsing and validation.
- `example/` — server adapters and example Provider.
- `reference/` — FDSN specification documents.

## Testing

```shell
$ go test -race ./...
$ go vet ./...
```

## Contributing

Contributions, issues, and pull requests are welcome.

## License

Licensed under the Apache License 2.0.
