package fdsnwsgo

import "github.com/bclswl0827/fdsnws-go/service"

type (
	Module                = service.Module
	Service               = service.Service
	Outcome               = service.Outcome
	RequestError          = service.RequestError
	Request               = service.Request
	Response              = service.Response
	PreparedResponse      = service.PreparedResponse
	AuthenticationRequest = service.AuthenticationRequest
	Principal             = service.Principal
	Authenticator         = service.Authenticator
	AuthenticatorFunc     = service.AuthenticatorFunc
	Result                = service.Result
	TimeRange             = service.TimeRange
	StreamEpoch           = service.StreamEpoch
	GeographicConstraints = service.GeographicConstraints
	DataSelectQuery       = service.DataSelectQuery
	StationQuery          = service.StationQuery
	EventQuery            = service.EventQuery
	DataSelectResult      = service.DataSelectResult
	DataRecord            = service.DataRecord
	CatalogResult         = service.CatalogResult
	ContributorResult     = service.ContributorResult
	UnimplementedProvider = service.UnimplementedProvider
	HTTPError             = service.HTTPError
	ProviderErrorKind     = service.ProviderErrorKind
)

type (
	IDataSelectProvider = service.IDataSelectProvider
	IStationProvider    = service.IStationProvider
	IEventProvider      = service.IEventProvider
	ITimeProvider       = service.ITimeProvider
	IRootProvider       = service.IRootProvider
	IProvider           = service.IProvider
)

const (
	ModuleDataSelect    = service.ModuleDataSelect
	ModuleStation       = service.ModuleStation
	ModuleEvent         = service.ModuleEvent
	MIMEText            = service.MIMEText
	MIMEXML             = service.MIMEXML
	MIMEMiniSEED        = service.MIMEMiniSEED
	ProviderErrorNone   = service.ProviderErrorNone
	ProviderErrorClient = service.ProviderErrorClient
	ProviderErrorServer = service.ProviderErrorServer
	defaultBasePath     = "/fdsnws"
)

var (
	ErrNoData         = service.ErrNoData
	ErrNotImplemented = service.ErrNotImplemented
)
