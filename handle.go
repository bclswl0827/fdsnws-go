package fdsnwsgo

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	component "github.com/bclswl0827/fdsnws-go/service"
)

type FDSNWS struct {
	services     map[Module]component.Service
	rootProvider component.IRootProvider
	timeProvider component.ITimeProvider
	config       config
}

type routeKind uint8

const (
	routeEndpoint routeKind = iota
	routeServiceRoot
	routeVersionRoot
)

func (f *FDSNWS) Handle(ctx context.Context, request Request) Response {
	if ctx == nil {
		ctx = context.Background()
	}
	request.ReceivedAt = f.timeProvider.Now().UTC()
	if len(request.URL) > 2000 || len(request.Path)+len(request.RawQuery)+1 > 2000 {
		return f.ErrorResponse(request, "", http.StatusRequestURITooLong, "Request URI too large", "Request URI exceeds 2000 bytes")
	}
	if strings.TrimSuffix(request.Path, "/") == f.config.basePath {
		if !strings.EqualFold(request.Method, http.MethodGet) {
			return f.ErrorResponse(request, "", http.StatusMethodNotAllowed, "Method not allowed", "This endpoint accepts GET")
		}
		if isNil(f.rootProvider) {
			return f.ErrorResponse(request, "", http.StatusNotFound, "Not found", "No FDSNWS root Provider is registered")
		}
		if err := ctx.Err(); err != nil {
			return f.finishOutcome(request, "", component.Outcome{}, err)
		}
		result, kind, err := f.rootProvider.FDSNWSRoot()
		outcome := preparedOutcome(result)
		outcome.ErrorKind = kind
		return f.finishOutcome(request, "", outcome, err)
	}
	service, endpoint, kind, ok := f.route(request.Path)
	if !ok {
		return f.ErrorResponse(request, "", http.StatusNotFound, "Not found", "Unknown or disabled FDSN web service endpoint")
	}
	module := service.Name()
	if kind == routeServiceRoot || kind == routeVersionRoot {
		if !strings.EqualFold(request.Method, http.MethodGet) {
			return f.ErrorResponse(request, module, http.StatusMethodNotAllowed, "Method not allowed", "This endpoint accepts GET")
		}
		if err := ctx.Err(); err != nil {
			return f.finishOutcome(request, module, component.Outcome{}, err)
		}
		outcome, err := service.Root(kind == routeVersionRoot)
		return f.finishOutcome(request, module, outcome, err)
	}
	if endpoint == "version" || endpoint == "application.wadl" {
		if !strings.EqualFold(request.Method, http.MethodGet) {
			return f.ErrorResponse(request, module, http.StatusMethodNotAllowed, "Method not allowed", "This endpoint accepts GET")
		}
		if endpoint == "version" {
			return Response{StatusCode: http.StatusOK, MIMEType: MIMEText, Body: []byte(f.serviceVersion(module) + "\n")}
		}
		return Response{StatusCode: http.StatusOK, MIMEType: MIMEXML, Body: []byte(service.WADL(f.config.basePath))}
	}

	outcome, err := service.Handle(ctx, endpoint, request)
	return f.finishOutcome(request, module, outcome, err)
}

func (f *FDSNWS) finishOutcome(request Request, module Module, outcome component.Outcome, err error) Response {
	if outcome.NoData == 0 {
		outcome.NoData = http.StatusNoContent
	}
	if err != nil {
		var requestErr *component.RequestError
		if errors.As(err, &requestErr) {
			response := f.ErrorResponse(request, module, requestErr.Status, requestErr.Summary, requestErr.Detail)
			response.Headers = cloneResponseHeaders(requestErr.Headers)
			return response
		}
		if errors.Is(err, ErrNoData) {
			return f.noDataResponse(request, module, outcome.NoData)
		}
		if errors.Is(err, ErrNotImplemented) {
			return f.ErrorResponse(request, module, http.StatusNotImplemented, "Not implemented", err.Error())
		}
		var httpErr *HTTPError
		if errors.As(err, &httpErr) && httpErr.Status >= 400 && httpErr.Status <= 599 {
			response := f.ErrorResponse(request, module, httpErr.Status, http.StatusText(httpErr.Status), httpErr.Error())
			response.Headers = cloneResponseHeaders(httpErr.Headers)
			return response
		}
		if outcome.ErrorKind == component.ProviderErrorClient {
			return f.ErrorResponse(request, module, http.StatusBadRequest, "Bad request", err.Error())
		}
		return f.ErrorResponse(request, module, http.StatusInternalServerError, "Internal server error", err.Error())
	}
	if isNil(outcome.Value) {
		return f.noDataResponse(request, module, outcome.NoData)
	}
	response, encodeErr := encodeResult(outcome, outcome.Value)
	if encodeErr != nil {
		return f.ErrorResponse(request, module, http.StatusInternalServerError, "Internal server error", encodeErr.Error())
	}
	return response
}

func preparedOutcome(result component.PreparedResponse) component.Outcome {
	return component.Outcome{Value: result.Body, MIMEType: result.MIMEType, StatusCode: result.StatusCode}
}

func (f *FDSNWS) route(path string) (component.Service, string, routeKind, bool) {
	path = strings.TrimSuffix(path, "/")
	prefix := f.config.basePath + "/"
	if !strings.HasPrefix(path, prefix) {
		return nil, "", routeEndpoint, false
	}
	parts := strings.Split(strings.TrimPrefix(path, prefix), "/")
	if len(parts) < 1 || len(parts) > 3 || parts[0] == "" {
		return nil, "", routeEndpoint, false
	}
	service, ok := f.services[Module(parts[0])]
	if !ok {
		return nil, "", routeEndpoint, false
	}
	if len(parts) == 1 {
		return service, "", routeServiceRoot, true
	}
	if parts[1] == "" || service.MajorVersion() != parts[1] {
		return nil, "", routeEndpoint, false
	}
	if len(parts) == 2 {
		return service, "", routeVersionRoot, true
	}
	if parts[2] == "" {
		return nil, "", routeEndpoint, false
	}
	return service, parts[2], routeEndpoint, true
}

func (f *FDSNWS) serviceVersion(module Module) string {
	if version, ok := f.config.versions[module]; ok {
		return version
	}
	if service, ok := f.services[module]; ok {
		return service.Version()
	}
	return "unknown"
}

func (f *FDSNWS) noDataResponse(request Request, module Module, status int) Response {
	if status == http.StatusNoContent {
		return Response{StatusCode: status, MIMEType: MIMEText}
	}
	return f.ErrorResponse(request, module, status, "No data", "The request was valid but matched no data")
}

// ErrorResponse builds an FDSN error response for use by middleware and adapters.
// It uses the configured documentation URL and the module's service version.
// Request.URL takes precedence over Request.Path and Request.RawQuery.
// If Request.ReceivedAt is zero, the Provider's clock supplies the submission time.
// Callers can set Response.Headers before writing with WriteHTTPResponse.
func (f *FDSNWS) ErrorResponse(request Request, module Module, status int, summary, detail string) Response {
	if request.ReceivedAt.IsZero() {
		request.ReceivedAt = f.timeProvider.Now()
	}
	docNotes := "Usage details are not provided by administrator"
	docLink := f.config.documentationURL
	if docLink != "" {
		docNotes = fmt.Sprintf("Usage details are available from %s", docLink)
	}

	submittedURL := request.URL
	if submittedURL == "" {
		submittedURL = request.Path
		if request.RawQuery != "" {
			submittedURL += "?" + request.RawQuery
		}
	}
	body := fmt.Sprintf("Error %d: %s\n\n%s\n\n%s\n\nRequest:\n%s\n\nRequest Submitted:\n%s\n\nService version:\n%s\n",
		status, summary, detail, docNotes, submittedURL, request.ReceivedAt.UTC().Format(time.RFC3339Nano), f.serviceVersion(module))
	return Response{StatusCode: status, MIMEType: MIMEText, Body: []byte(body)}
}

func cloneResponseHeaders(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	result := make(map[string]string, len(input))
	for key, value := range input {
		result[key] = value
	}
	return result
}
