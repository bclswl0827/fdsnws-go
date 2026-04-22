package fdsnwsgo

import (
	"context"
	"fmt"
	"io"
	"net/http"
)

// RequestFromHTTP normalizes a standard library request. Other frameworks can
// construct Request directly without depending on this adapter.
func RequestFromHTTP(request *http.Request) (Request, error) {
	if request == nil {
		return Request{}, fmt.Errorf("HTTP request cannot be nil")
	}
	body, err := io.ReadAll(request.Body)
	if err != nil {
		return Request{}, fmt.Errorf("read request body: %w", err)
	}
	headers := make(map[string][]string, len(request.Header))
	for key, values := range request.Header {
		headers[key] = append([]string(nil), values...)
	}
	return Request{
		Method: request.Method, Path: request.URL.Path, RawQuery: request.URL.RawQuery,
		Query: request.URL.Query(), Headers: headers, Body: body,
		RemoteAddr: request.RemoteAddr, URL: request.URL.String(),
	}, nil
}

// WriteHTTPResponse writes a normalized response through net/http.
func WriteHTTPResponse(writer http.ResponseWriter, response Response) error {
	if writer == nil {
		return fmt.Errorf("HTTP response writer cannot be nil")
	}
	for key, value := range response.Headers {
		writer.Header().Set(key, value)
	}
	if response.MIMEType != "" {
		writer.Header().Set("Content-Type", response.MIMEType)
	}
	writer.WriteHeader(response.StatusCode)
	if len(response.Body) == 0 {
		return nil
	}
	_, err := writer.Write(response.Body)
	return err
}

// ServeHTTP is an optional net/http adapter. The protocol core remains usable
// with any server framework through Handle.
func (f *FDSNWS) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	normalized, err := RequestFromHTTP(request)
	if err != nil {
		response := f.errorResponse(Request{Path: request.URL.Path, URL: request.URL.String()}, "", http.StatusBadRequest, "Bad request", err.Error())
		_ = WriteHTTPResponse(writer, response)
		return
	}
	ctx := request.Context()
	if ctx == nil {
		ctx = context.Background()
	}
	_ = WriteHTTPResponse(writer, f.Handle(ctx, normalized))
}
