// Package service defines the common contract implemented by every FDSN web
// service component.
package service

import (
	"context"
	"net/url"
	"time"

	"github.com/bclswl0827/fdsnws-go/quakeml"
	"github.com/bclswl0827/fdsnws-go/stationxml"
)

type Module string

const (
	ModuleDataSelect Module = "dataselect"
	ModuleStation    Module = "station"
	ModuleEvent      Module = "event"
)

const (
	MIMEText     = "text/plain; charset=utf-8"
	MIMEXML      = "application/xml; charset=utf-8"
	MIMEMiniSEED = "application/vnd.fdsn.mseed"
)

// Request is the normalized, server-framework-independent request.
type Request struct {
	Method     string
	Path       string
	RawQuery   string
	Query      url.Values
	Headers    map[string][]string
	Body       []byte
	RemoteAddr string
	URL        string
	ReceivedAt time.Time
}

// AuthenticationRequest contains only the transport data needed to validate
// credentials. It deliberately excludes the server framework context and the
// parsed FDSN query.
type AuthenticationRequest struct {
	Method        string
	RequestTarget string
	Authorization string
	Body          []byte
	RemoteAddr    string
}

// Principal is the authenticated identity exposed to a data Provider. The
// Provider never receives credentials or the transport request itself.
type Principal struct {
	Subject    string
	Attributes map[string]string
}

// Authenticator validates queryauth credentials. Invalid credentials should
// normally be reported with an HTTPError containing status 401 and the
// appropriate WWW-Authenticate challenge.
type Authenticator interface {
	Authenticate(context.Context, AuthenticationRequest) (Principal, error)
}

type AuthenticatorFunc func(context.Context, AuthenticationRequest) (Principal, error)

func (f AuthenticatorFunc) Authenticate(ctx context.Context, request AuthenticationRequest) (Principal, error) {
	return f(ctx, request)
}

type Response struct {
	StatusCode int
	MIMEType   string
	Headers    map[string]string
	Body       []byte
}

// Result is an optional response envelope for custom Service implementations.
type Result struct {
	Body     any
	MIMEType string
	Headers  map[string]string
}

// Outcome is the normalized result returned by a Service component to the
// protocol router.
type Outcome struct {
	Value      any
	StatusCode int
	ErrorKind  ProviderErrorKind
	Format     string
	MIMEType   string
	XMLRoot    string
	ListRoot   string
	ListItem   string
	NoData     int
}

// PreparedResponse is a transport-neutral, prebuilt response returned by a
// Provider for an FDSNWS or service landing path.
type PreparedResponse struct {
	Body       string
	MIMEType   string
	StatusCode int
}

// Service is the extension point for an FDSN service. Implementations own
// their endpoints, request parsing, validation, Provider invocation and WADL.
// The root router owns the common /version and /application.wadl behavior.
type Service interface {
	Name() Module
	MajorVersion() string
	Version() string
	Root(bool) (Outcome, error)
	Handle(context.Context, string, Request) (Outcome, error)
	WADL(string) string
}

// RequestError describes a client or routing error detected by a component.
type RequestError struct {
	Status  int
	Summary string
	Detail  string
	Headers map[string]string
}

func (e *RequestError) Error() string { return e.Detail }

type TimeRange struct {
	Start *time.Time
	End   *time.Time
}

type StreamEpoch struct {
	Network  string
	Station  string
	Location string
	Channel  string
	TimeRange
}

type GeographicConstraints struct {
	MinLatitude  *float64
	MaxLatitude  *float64
	MinLongitude *float64
	MaxLongitude *float64
	Latitude     *float64
	Longitude    *float64
	MinRadius    *float64
	MaxRadius    *float64
}

type DataSelectQuery struct {
	Selections    []StreamEpoch
	Quality       string
	MinimumLength float64
	LongestOnly   bool
	Format        string
	NoData        int
	Principal     *Principal
}

type StationQuery struct {
	Selections          []StreamEpoch
	StartBefore         *time.Time
	StartAfter          *time.Time
	EndBefore           *time.Time
	EndAfter            *time.Time
	Geographic          GeographicConstraints
	Level               string
	IncludeRestricted   bool
	IncludeAvailability bool
	UpdatedAfter        *time.Time
	MatchTimeSeries     bool
	Format              string
	NoData              int
}

type EventQuery struct {
	TimeRange            TimeRange
	Geographic           GeographicConstraints
	MinDepth             *float64
	MaxDepth             *float64
	MinMagnitude         *float64
	MaxMagnitude         *float64
	MagnitudeType        string
	EventType            []string
	IncludeAllOrigins    bool
	IncludeAllMagnitudes bool
	IncludeArrivals      bool
	EventID              string
	Limit                *int
	Offset               int
	OrderBy              string
	Contributor          string
	Catalog              string
	UpdatedAfter         *time.Time
	Format               string
	NoData               int
}

type IDataSelectProvider interface {
	DataSelectRoot() (PreparedResponse, ProviderErrorKind, error)
	DataSelectVersionRoot() (PreparedResponse, ProviderErrorKind, error)
	DataSelect(DataSelectQuery) (DataSelectResult, ProviderErrorKind, error)
}

// ITimeProvider is the single source of wall-clock time used by the protocol
// layer. Providers may return a fixed time for deterministic deployments and
// tests.
type ITimeProvider interface {
	Now() time.Time
}

type IStationProvider interface {
	StationRoot() (PreparedResponse, ProviderErrorKind, error)
	StationVersionRoot() (PreparedResponse, ProviderErrorKind, error)
	Station(StationQuery) (*stationxml.Document, ProviderErrorKind, error)
}

type IEventProvider interface {
	EventRoot() (PreparedResponse, ProviderErrorKind, error)
	EventVersionRoot() (PreparedResponse, ProviderErrorKind, error)
	Event(EventQuery) (*quakeml.Document, ProviderErrorKind, error)
	EventCatalogs() (CatalogResult, ProviderErrorKind, error)
	EventContributors() (ContributorResult, ProviderErrorKind, error)
}

type IRootProvider interface {
	FDSNWSRoot() (PreparedResponse, ProviderErrorKind, error)
}

type IProvider interface {
	ITimeProvider
	IRootProvider
	IDataSelectProvider
	IStationProvider
	IEventProvider
}
