package service

import (
	"errors"
	"fmt"
)

var (
	ErrNoData         = errors.New("no data")
	ErrNotImplemented = errors.New("provider method not implemented")
)

// ProviderErrorKind classifies an ordinary Provider error by who can correct
// it. Exact protocol errors may still use HTTPError.
type ProviderErrorKind uint8

const (
	ProviderErrorNone ProviderErrorKind = iota
	ProviderErrorClient
	ProviderErrorServer
)

type HTTPError struct {
	Status  int
	Detail  string
	Err     error
	Headers map[string]string
}

func (e *HTTPError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("HTTP status %d", e.Status)
}

func (e *HTTPError) Unwrap() error { return e.Err }
