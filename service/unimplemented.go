package service

import (
	"time"

	"github.com/bclswl0827/fdsnws-go/quakeml"
	"github.com/bclswl0827/fdsnws-go/stationxml"
)

type UnimplementedProvider struct{}

func (UnimplementedProvider) Now() time.Time { return time.Now().UTC() }

func (UnimplementedProvider) FDSNWSRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) DataSelectRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) DataSelectVersionRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) DataSelect(DataSelectQuery) (DataSelectResult, ProviderErrorKind, error) {
	return DataSelectResult{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) StationRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) StationVersionRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) Station(StationQuery) (*stationxml.Document, ProviderErrorKind, error) {
	return nil, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) EventRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) EventVersionRoot() (PreparedResponse, ProviderErrorKind, error) {
	return PreparedResponse{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) Event(EventQuery) (*quakeml.Document, ProviderErrorKind, error) {
	return nil, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) EventCatalogs() (CatalogResult, ProviderErrorKind, error) {
	return CatalogResult{}, ProviderErrorServer, ErrNotImplemented
}
func (UnimplementedProvider) EventContributors() (ContributorResult, ProviderErrorKind, error) {
	return ContributorResult{}, ProviderErrorServer, ErrNotImplemented
}
