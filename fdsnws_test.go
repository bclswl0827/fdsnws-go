package fdsnwsgo

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	quakemlformat "github.com/bclswl0827/fdsnws-go/quakeml"
	component "github.com/bclswl0827/fdsnws-go/service"
	stationcomponent "github.com/bclswl0827/fdsnws-go/service/station"
	"github.com/bclswl0827/mseedio"
	stationxmlformat "github.com/bclswl0827/obsgo/inventory/stationxml"
)

type testProvider struct {
	UnimplementedProvider
	currentTime           time.Time
	errorKind             ProviderErrorKind
	root                  func() (PreparedResponse, error)
	dataSelectRoot        func() (PreparedResponse, error)
	dataSelectVersionRoot func() (PreparedResponse, error)
	stationRoot           func() (PreparedResponse, error)
	stationVersionRoot    func() (PreparedResponse, error)
	eventRoot             func() (PreparedResponse, error)
	eventVersionRoot      func() (PreparedResponse, error)
	dataSelect            func(DataSelectQuery) (DataSelectResult, error)
	station               func(StationQuery) (*stationxmlformat.Document, error)
	event                 func(EventQuery) (*quakemlformat.Document, error)
	catalogs              func() (CatalogResult, error)
	contributors          func() (ContributorResult, error)
}

func (p testProvider) Now() time.Time {
	if !p.currentTime.IsZero() {
		return p.currentTime
	}
	return p.UnimplementedProvider.Now()
}

func (p testProvider) FDSNWSRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.root == nil {
		return p.UnimplementedProvider.FDSNWSRoot()
	}
	result, err := p.root()
	return result, p.errorKind, err
}

func (p testProvider) DataSelectRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.dataSelectRoot == nil {
		return p.UnimplementedProvider.DataSelectRoot()
	}
	result, err := p.dataSelectRoot()
	return result, p.errorKind, err
}

func (p testProvider) DataSelectVersionRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.dataSelectVersionRoot == nil {
		return p.UnimplementedProvider.DataSelectVersionRoot()
	}
	result, err := p.dataSelectVersionRoot()
	return result, p.errorKind, err
}

func (p testProvider) StationRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.stationRoot == nil {
		return p.UnimplementedProvider.StationRoot()
	}
	result, err := p.stationRoot()
	return result, p.errorKind, err
}

func (p testProvider) StationVersionRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.stationVersionRoot == nil {
		return p.UnimplementedProvider.StationVersionRoot()
	}
	result, err := p.stationVersionRoot()
	return result, p.errorKind, err
}

func (p testProvider) EventRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.eventRoot == nil {
		return p.UnimplementedProvider.EventRoot()
	}
	result, err := p.eventRoot()
	return result, p.errorKind, err
}

func (p testProvider) EventVersionRoot() (PreparedResponse, ProviderErrorKind, error) {
	if p.eventVersionRoot == nil {
		return p.UnimplementedProvider.EventVersionRoot()
	}
	result, err := p.eventVersionRoot()
	return result, p.errorKind, err
}

func (p testProvider) DataSelect(query DataSelectQuery) (DataSelectResult, ProviderErrorKind, error) {
	if p.dataSelect == nil {
		return p.UnimplementedProvider.DataSelect(query)
	}
	result, err := p.dataSelect(query)
	return result, p.errorKind, err
}

func (p testProvider) Station(query StationQuery) (*stationxmlformat.Document, ProviderErrorKind, error) {
	if p.station == nil {
		return p.UnimplementedProvider.Station(query)
	}
	result, err := p.station(query)
	return result, p.errorKind, err
}

func (p testProvider) Event(query EventQuery) (*quakemlformat.Document, ProviderErrorKind, error) {
	if p.event == nil {
		return p.UnimplementedProvider.Event(query)
	}
	result, err := p.event(query)
	return result, p.errorKind, err
}

func (p testProvider) EventCatalogs() (CatalogResult, ProviderErrorKind, error) {
	if p.catalogs == nil {
		return p.UnimplementedProvider.EventCatalogs()
	}
	result, err := p.catalogs()
	return result, p.errorKind, err
}

func (p testProvider) EventContributors() (ContributorResult, ProviderErrorKind, error) {
	if p.contributors == nil {
		return p.UnimplementedProvider.EventContributors()
	}
	result, err := p.contributors()
	return result, p.errorKind, err
}

func normalizedRequest(method, path, rawQuery, body string) Request {
	values, _ := url.ParseQuery(rawQuery)
	requestURL := path
	if rawQuery != "" {
		requestURL += "?" + rawQuery
	}
	return Request{Method: method, Path: path, RawQuery: rawQuery, Query: values, Body: []byte(body), URL: requestURL}
}

func waveformResult() DataSelectResult {
	return DataSelectResult{Records: []DataRecord{{
		Network: "IU", Station: "ANMO", Location: "00", Channel: "BHZ", Quality: "D",
		StartTime: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), SampleRate: 20,
		Samples: []int32{1, 2, 3, 4, 5, 6, 7, 8},
	}}}
}

func stationResult() *stationxmlformat.Document {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	document := stationxmlformat.NewDocument()
	document.Source = "unit-test"
	document.Created = start
	document.Networks = []stationxmlformat.Network{{
		Code: "XX", Description: "Test network", StartDate: start,
		Stations: []stationxmlformat.Station{{
			Code: "TEST", Latitude: 1, Longitude: 2, Elevation: 3,
			Site: stationxmlformat.Site{Name: "Test station"}, StartDate: start,
			Channels: []stationxmlformat.Channel{{
				Code: "BHZ", LocationCode: "00", Latitude: 1, Longitude: 2, Elevation: 3,
				Depth: 0, Azimuth: 0, Dip: -90, SampleRate: 20, StartDate: start,
				Response: &stationxmlformat.Response{Sensitivity: &stationxmlformat.InstrumentSensitivity{
					Value: 2, Frequency: 1,
					Input: stationxmlformat.Units{Name: "m/s"}, Output: stationxmlformat.Units{Name: "count"},
				}},
			}},
		}},
	}}
	return document
}

func eventResult() *quakemlformat.Document {
	originTime := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	document := quakemlformat.NewDocument("smi:local/unit-test")
	document.Parameters.Events = []quakemlformat.Event{{
		PublicID: "smi:local/event-1", Type: "earthquake",
		Description:       &quakemlformat.Description{Text: "Test event", Type: "earthquake name"},
		PreferredOriginID: "smi:local/origin-1", PreferredMagnitudeID: "smi:local/magnitude-1",
		CreationInfo: &quakemlformat.CreationInfo{AgencyID: "NEIC"},
		Origins: []quakemlformat.Origin{{
			PublicID: "smi:local/origin-1", Time: quakemlformat.TimeQuantity{Value: originTime.Format(time.RFC3339Nano)},
			Latitude: quakemlformat.RealQuantity{Value: 1}, Longitude: quakemlformat.RealQuantity{Value: 2},
			Depth: quakemlformat.RealQuantity{Value: 3000}, CreationInfo: &quakemlformat.CreationInfo{Author: "test"},
		}},
		Magnitudes: []quakemlformat.Magnitude{{
			PublicID: "smi:local/magnitude-1", OriginID: "smi:local/origin-1", Type: "Mw",
			Mag: quakemlformat.RealQuantity{Value: 4.5}, CreationInfo: &quakemlformat.CreationInfo{Author: "test"},
		}},
	}}
	return document
}

func TestDataSelectGETParsesAliasesAndEncodesMiniSEED(t *testing.T) {
	called := false
	provider := testProvider{dataSelect: func(query DataSelectQuery) (DataSelectResult, error) {
		called = true
		if len(query.Selections) != 1 {
			t.Fatalf("selections = %d", len(query.Selections))
		}
		selection := query.Selections[0]
		if selection.Network != "IU" || selection.Location != "  " {
			t.Fatalf("selection = %#v", selection)
		}
		if query.Quality != "B" || query.Format != "miniseed" || query.NoData != 204 {
			t.Fatalf("defaults = %#v", query)
		}
		return waveformResult(), nil
	}}
	server, err := NewFDSNWS(provider, WithModules(ModuleDataSelect))
	if err != nil {
		t.Fatal(err)
	}
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/query", "net=IU&sta=ANMO&loc=--&cha=BHZ&start=2020-01-01&end=2020-01-02", ""))
	if !called {
		t.Fatal("provider was not called")
	}
	if response.StatusCode != http.StatusOK || response.MIMEType != MIMEMiniSEED || len(response.Body) == 0 {
		t.Fatalf("response = %#v", response)
	}
	var decoded mseedio.MiniSeedData
	if err := decoded.ReadFromReader(bytes.NewReader(response.Body)); err != nil {
		t.Fatalf("decode miniSEED: %v", err)
	}
	if decoded.Records != 1 || decoded.Samples != 8 {
		t.Fatalf("decoded = %#v", decoded)
	}
}

func TestDataSelectPOSTMultipleSelections(t *testing.T) {
	provider := testProvider{dataSelect: func(query DataSelectQuery) (DataSelectResult, error) {
		if len(query.Selections) != 2 {
			t.Fatalf("selections = %d", len(query.Selections))
		}
		if query.Quality != "M" || !query.LongestOnly {
			t.Fatalf("query = %#v", query)
		}
		result := waveformResult()
		result.Records[0].Quality = "M"
		return result, nil
	}}
	server, _ := NewFDSNWS(provider, WithModules(ModuleDataSelect))
	body := "quality=M\nlongestonly=TRUE\nIU ANMO 00 BHZ 2020-01-01 2020-01-02\nIU COLA -- HH? 2020-01-03 2020-01-04\n"
	response := server.Handle(context.Background(), normalizedRequest(http.MethodPost, "/fdsnws/dataselect/1/query", "", body))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("response = %#v", response)
	}
}

func TestDataSelectQueryAuthIsDisabledWithoutAuthenticator(t *testing.T) {
	providerCalled := false
	provider := testProvider{dataSelect: func(DataSelectQuery) (DataSelectResult, error) {
		providerCalled = true
		return waveformResult(), nil
	}}
	server, _ := NewFDSNWS(provider, WithModules(ModuleDataSelect))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/queryauth", "", ""))
	if response.StatusCode != http.StatusNotFound || providerCalled {
		t.Fatalf("response = %#v, providerCalled = %t", response, providerCalled)
	}
	wadl := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/application.wadl", "", ""))
	if strings.Contains(string(wadl.Body), `resource path="queryauth"`) {
		t.Fatalf("disabled queryauth appears in WADL: %s", wadl.Body)
	}
}

func TestDataSelectQueryAuthUsesIndependentAuthenticator(t *testing.T) {
	const rawQuery = "net=IU&sta=ANMO&loc=00&cha=BHZ&start=2020-01-01&end=2020-01-02"
	authenticatorCalled := false
	authenticator := AuthenticatorFunc(func(_ context.Context, request AuthenticationRequest) (Principal, error) {
		authenticatorCalled = true
		if request.Method != http.MethodGet || request.RequestTarget != "/fdsnws/dataselect/1/queryauth?"+rawQuery || request.Authorization != "Digest token" {
			t.Fatalf("authentication request = %#v", request)
		}
		return Principal{Subject: "station-user", Attributes: map[string]string{"scope": "restricted"}}, nil
	})
	provider := testProvider{dataSelect: func(query DataSelectQuery) (DataSelectResult, error) {
		if query.Principal == nil || query.Principal.Subject != "station-user" || query.Principal.Attributes["scope"] != "restricted" {
			t.Fatalf("principal = %#v", query.Principal)
		}
		return waveformResult(), nil
	}}
	server, err := NewFDSNWS(provider, WithModules(ModuleDataSelect), WithAuthenticator(authenticator))
	if err != nil {
		t.Fatal(err)
	}
	request := normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/queryauth", rawQuery, "")
	request.Headers = map[string][]string{"authorization": {"Digest token"}}
	response := server.Handle(context.Background(), request)
	if response.StatusCode != http.StatusOK || !authenticatorCalled {
		t.Fatalf("response = %#v, authenticatorCalled = %t", response, authenticatorCalled)
	}
	wadl := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/application.wadl", "", ""))
	if !strings.Contains(string(wadl.Body), `resource path="queryauth"`) {
		t.Fatalf("enabled queryauth missing from WADL: %s", wadl.Body)
	}
}

func TestDataSelectPublicQueryDoesNotUseAuthenticator(t *testing.T) {
	authenticator := AuthenticatorFunc(func(context.Context, AuthenticationRequest) (Principal, error) {
		t.Fatal("public query called authenticator")
		return Principal{}, nil
	})
	provider := testProvider{dataSelect: func(query DataSelectQuery) (DataSelectResult, error) {
		if query.Principal != nil {
			t.Fatalf("public query principal = %#v", query.Principal)
		}
		return waveformResult(), nil
	}}
	server, _ := NewFDSNWS(provider, WithModules(ModuleDataSelect), WithAuthenticator(authenticator))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/query", "net=IU&sta=ANMO&loc=00&cha=BHZ&start=2020-01-01&end=2020-01-02", ""))
	if response.StatusCode != http.StatusOK {
		t.Fatalf("response = %#v", response)
	}
}

func TestDataSelectQueryAuthPreservesAuthenticationChallenge(t *testing.T) {
	authenticator := AuthenticatorFunc(func(context.Context, AuthenticationRequest) (Principal, error) {
		return Principal{}, &HTTPError{Status: http.StatusUnauthorized, Detail: "authentication required", Headers: map[string]string{"WWW-Authenticate": `Digest realm="FDSN"`}}
	})
	server, _ := NewFDSNWS(testProvider{}, WithModules(ModuleDataSelect), WithAuthenticator(authenticator))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/dataselect/1/queryauth", "", ""))
	if response.StatusCode != http.StatusUnauthorized || response.Headers["WWW-Authenticate"] != `Digest realm="FDSN"` {
		t.Fatalf("response = %#v", response)
	}
}

func TestStationTypedResultEncodesResponseXML(t *testing.T) {
	provider := testProvider{station: func(query StationQuery) (*stationxmlformat.Document, error) {
		if query.Level != "response" || query.Format != "xml" {
			t.Fatalf("query = %#v", query)
		}
		return stationResult(), nil
	}}
	server, _ := NewFDSNWS(provider, WithModules(ModuleStation))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "level=response", ""))
	if response.StatusCode != http.StatusOK || response.MIMEType != MIMEXML {
		t.Fatalf("response = %#v", response)
	}
	body := string(response.Body)
	for _, fragment := range []string{`<FDSNStationXML`, `<Network code="XX"`, `<Response>`, `<InstrumentSensitivity>`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("missing %q in %s", fragment, body)
		}
	}
	document, err := stationxmlformat.Parse(response.Body)
	if err != nil || document.Networks[0].Stations[0].Channels[0].Response.Sensitivity.Value != 2 {
		t.Fatalf("parse StationXML: document = %#v, error = %v", document, err)
	}
}

func TestStationTypedResultEncodesText(t *testing.T) {
	provider := testProvider{station: func(StationQuery) (*stationxmlformat.Document, error) { return stationResult(), nil }}
	server, _ := NewFDSNWS(provider, WithModules(ModuleStation))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "level=channel&format=text", ""))
	if response.StatusCode != http.StatusOK || response.MIMEType != MIMEText || !strings.Contains(string(response.Body), "XX|TEST|00|BHZ") {
		t.Fatalf("response = %#v", response)
	}
}

func TestProviderTimeIsUsedForRequestsAndStationXML(t *testing.T) {
	fixed := time.Date(2030, 4, 5, 6, 7, 8, 9, time.FixedZone("test", 8*60*60))
	document := stationResult()
	document.Created = ""
	provider := testProvider{
		currentTime: fixed,
		station: func(StationQuery) (*stationxmlformat.Document, error) {
			return document, nil
		},
	}
	server, _ := NewFDSNWS(provider, WithModules(ModuleStation))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "level=network", ""))
	decoded, err := stationxmlformat.Parse(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	want := fixed.UTC().Format(time.RFC3339Nano)
	if decoded.Created != want {
		t.Fatalf("StationXML Created = %q, want %q", decoded.Created, want)
	}

	request := normalizedRequest(http.MethodGet, "/not-found", "", "")
	request.ReceivedAt = time.Date(1999, 1, 1, 0, 0, 0, 0, time.UTC)
	errorResponse := server.Handle(context.Background(), request)
	if !strings.Contains(string(errorResponse.Body), "Request Submitted:\n"+want) {
		t.Fatalf("error response does not use Provider time: %s", errorResponse.Body)
	}
}

func TestStationRejectsTextResponseLevel(t *testing.T) {
	server, _ := NewFDSNWS(testProvider{}, WithModules(ModuleStation))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "level=response&format=text", ""))
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(response.Body), "format=text") {
		t.Fatalf("response = %#v", response)
	}
}

func TestStationRejectsMixedTimeConstraintGroups(t *testing.T) {
	server, _ := NewFDSNWS(testProvider{}, WithModules(ModuleStation))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "starttime=2020-01-01&startbefore=2021-01-01", ""))
	if response.StatusCode != http.StatusBadRequest || !strings.Contains(string(response.Body), "mutually exclusive") {
		t.Fatalf("response = %#v", response)
	}
}

func TestEventTypedResultAndAuxiliaryMethods(t *testing.T) {
	provider := testProvider{
		event: func(query EventQuery) (*quakemlformat.Document, error) {
			if query.Offset != 2 || query.OrderBy != "magnitude-asc" || len(query.EventType) != 2 {
				t.Fatalf("query = %#v", query)
			}
			return eventResult(), nil
		},
		catalogs:     func() (CatalogResult, error) { return CatalogResult{Catalogs: []string{"GCMT", "ISC"}}, nil },
		contributors: func() (ContributorResult, error) { return ContributorResult{Contributors: []string{"NEIC"}}, nil },
	}
	server, _ := NewFDSNWS(provider, WithModules(ModuleEvent))
	queryResponse := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "format=text&offset=2&orderby=magnitude-asc&eventtype=earthquake,explosion", ""))
	if queryResponse.StatusCode != http.StatusOK || queryResponse.MIMEType != MIMEText || !strings.Contains(string(queryResponse.Body), "smi:local/event-1") {
		t.Fatalf("query response = %#v", queryResponse)
	}
	catalogResponse := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/catalogs", "", ""))
	if catalogResponse.StatusCode != http.StatusOK || !strings.Contains(string(catalogResponse.Body), "<Catalog>GCMT</Catalog>") {
		t.Fatalf("catalog response = %#v", catalogResponse)
	}
	contributorResponse := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/contributors", "", ""))
	if !strings.Contains(string(contributorResponse.Body), "<Contributor>NEIC</Contributor>") {
		t.Fatalf("contributor response = %#v", contributorResponse)
	}
}

func TestEventTypedResultEncodesQuakeML(t *testing.T) {
	provider := testProvider{event: func(EventQuery) (*quakemlformat.Document, error) { return eventResult(), nil }}
	server, _ := NewFDSNWS(provider, WithModules(ModuleEvent))
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "", ""))
	if response.StatusCode != http.StatusOK || response.MIMEType != MIMEXML {
		t.Fatalf("response = %#v", response)
	}
	for _, fragment := range []string{"<quakeml", "<eventParameters", `<event publicID="smi:local/event-1"`, "<magnitude"} {
		if !strings.Contains(string(response.Body), fragment) {
			t.Fatalf("missing %q in %s", fragment, response.Body)
		}
	}
	document, err := quakemlformat.Parse(response.Body)
	if err != nil || document.Parameters.Events[0].PublicID != "smi:local/event-1" {
		t.Fatalf("parse QuakeML: document = %#v, error = %v", document, err)
	}
}

func TestServiceProjectionDoesNotMutateProviderDocuments(t *testing.T) {
	stationDocument := stationResult()
	stationProvider := testProvider{station: func(StationQuery) (*stationxmlformat.Document, error) {
		return stationDocument, nil
	}}
	stationServer, _ := NewFDSNWS(stationProvider, WithModules(ModuleStation))
	stationResponse := stationServer.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "level=channel", ""))
	projectedStation, err := stationxmlformat.Parse(stationResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	if projectedStation.Networks[0].Stations[0].Channels[0].Response != nil {
		t.Fatal("channel-level response contains response metadata")
	}
	if stationDocument.Networks[0].Stations[0].Channels[0].Response == nil {
		t.Fatal("station service mutated the provider document")
	}

	eventDocument := eventResult()
	eventDocument.Parameters.Events[0].Origins = append(eventDocument.Parameters.Events[0].Origins, quakemlformat.Origin{PublicID: "smi:local/origin-2"})
	eventProvider := testProvider{event: func(EventQuery) (*quakemlformat.Document, error) {
		return eventDocument, nil
	}}
	eventServer, _ := NewFDSNWS(eventProvider, WithModules(ModuleEvent))
	eventResponse := eventServer.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "", ""))
	projectedEvent, err := quakemlformat.Parse(eventResponse.Body)
	if err != nil {
		t.Fatal(err)
	}
	if len(projectedEvent.Parameters.Events[0].Origins) != 1 {
		t.Fatalf("projected origins = %d", len(projectedEvent.Parameters.Events[0].Origins))
	}
	if len(eventDocument.Parameters.Events[0].Origins) != 2 {
		t.Fatal("event service mutated the provider document")
	}
}

func TestNoDataStatusAndProviderHTTPError(t *testing.T) {
	provider := testProvider{event: func(query EventQuery) (*quakemlformat.Document, error) {
		if query.EventID == "forbidden" {
			return nil, &HTTPError{Status: http.StatusForbidden, Detail: "restricted catalog", Headers: map[string]string{"X-Reason": "policy"}}
		}
		return nil, ErrNoData
	}}
	server, _ := NewFDSNWS(provider, WithModules(ModuleEvent))
	noData := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "nodata=404", ""))
	if noData.StatusCode != http.StatusNotFound || !strings.HasPrefix(string(noData.Body), "Error 404") {
		t.Fatalf("response = %#v", noData)
	}
	forbidden := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "eventid=forbidden", ""))
	if forbidden.StatusCode != http.StatusForbidden || !strings.Contains(string(forbidden.Body), "restricted catalog") || forbidden.Headers["X-Reason"] != "policy" {
		t.Fatalf("response = %#v", forbidden)
	}
}

func TestProviderErrorKindControlsHTTPErrorClass(t *testing.T) {
	providerError := errors.New("provider failed")
	clientProvider := testProvider{
		errorKind: ProviderErrorClient,
		event:     func(EventQuery) (*quakemlformat.Document, error) { return nil, providerError },
	}
	clientServer, _ := NewFDSNWS(clientProvider, WithModules(ModuleEvent))
	clientResponse := clientServer.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "", ""))
	if clientResponse.StatusCode != http.StatusBadRequest || !strings.Contains(string(clientResponse.Body), providerError.Error()) {
		t.Fatalf("client response = %#v", clientResponse)
	}

	serverProvider := testProvider{
		errorKind: ProviderErrorServer,
		event:     func(EventQuery) (*quakemlformat.Document, error) { return nil, providerError },
	}
	server, _ := NewFDSNWS(serverProvider, WithModules(ModuleEvent))
	serverResponse := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", "", ""))
	if serverResponse.StatusCode != http.StatusInternalServerError || !strings.Contains(string(serverResponse.Body), providerError.Error()) {
		t.Fatalf("server response = %#v", serverResponse)
	}
}

func TestModuleSelectionVersionAndWADL(t *testing.T) {
	provider := testProvider{stationRoot: func() (PreparedResponse, error) {
		return PreparedResponse{Body: "station landing", MIMEType: "text/x-station", StatusCode: http.StatusAccepted}, nil
	}, stationVersionRoot: func() (PreparedResponse, error) {
		return PreparedResponse{Body: "station v1 landing", MIMEType: "text/x-station-v1", StatusCode: http.StatusNonAuthoritativeInfo}, nil
	}}
	server, err := NewFDSNWS(provider, WithModules(ModuleStation), WithServiceVersion(ModuleStation, "1.1.9"))
	if err != nil {
		t.Fatal(err)
	}
	disabled := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event", "", ""))
	if disabled.StatusCode != http.StatusNotFound {
		t.Fatalf("disabled response = %#v", disabled)
	}
	version := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/version", "", ""))
	if string(version.Body) != "1.1.9\n" {
		t.Fatalf("version = %q", version.Body)
	}
	for _, path := range []string{"/fdsnws/station", "/fdsnws/station/"} {
		landing := server.Handle(context.Background(), normalizedRequest(http.MethodGet, path, "ignored=yes", ""))
		if landing.StatusCode != http.StatusAccepted || landing.MIMEType != "text/x-station" || string(landing.Body) != "station landing" {
			t.Fatalf("path %s: landing = %#v", path, landing)
		}
	}
	for _, path := range []string{"/fdsnws/station/1", "/fdsnws/station/1/"} {
		landing := server.Handle(context.Background(), normalizedRequest(http.MethodGet, path, "ignored=yes", ""))
		if landing.StatusCode != http.StatusNonAuthoritativeInfo || landing.MIMEType != "text/x-station-v1" || string(landing.Body) != "station v1 landing" {
			t.Fatalf("path %s: landing = %#v", path, landing)
		}
	}
	wadl := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/application.wadl", "", ""))
	if wadl.StatusCode != http.StatusOK || wadl.MIMEType != MIMEXML || !strings.Contains(string(wadl.Body), `name="includeavailability"`) {
		t.Fatalf("wadl = %#v", wadl)
	}
}

func TestDataSelectBasePathReturnsPreparedProviderResponse(t *testing.T) {
	provider := testProvider{dataSelectRoot: func() (PreparedResponse, error) {
		return PreparedResponse{Body: "<h1>dataselect</h1>", MIMEType: "text/html; charset=utf-8", StatusCode: http.StatusAccepted}, nil
	}}
	server, _ := NewFDSNWS(provider, WithModules(ModuleDataSelect))
	request := httptest.NewRequest(http.MethodGet, "/fdsnws/dataselect", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusAccepted || recorder.Header().Get("Content-Type") != "text/html; charset=utf-8" || recorder.Body.String() != "<h1>dataselect</h1>" {
		t.Fatalf("status = %d, headers = %#v, body = %s", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}

func TestFDSNWSBasePathReturnsPreparedProviderResponse(t *testing.T) {
	provider := testProvider{root: func() (PreparedResponse, error) {
		return PreparedResponse{Body: "FDSNWS landing", MIMEType: "text/x-fdsnws", StatusCode: http.StatusCreated}, nil
	}}
	server, _ := NewFDSNWS(provider, WithModules())
	for _, path := range []string{"/fdsnws", "/fdsnws/"} {
		response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, path, "", ""))
		if response.StatusCode != http.StatusCreated || response.MIMEType != "text/x-fdsnws" || string(response.Body) != "FDSNWS landing" {
			t.Fatalf("path %s: response = %#v", path, response)
		}
	}
}

func TestValidation(t *testing.T) {
	server, _ := NewFDSNWS(testProvider{}, WithModules(ModuleEvent))
	tests := []string{
		"minlatitude=-91", "latitude=1", "latitude=1&longitude=2&minlatitude=0",
		"minmagnitude=1e2", "limit=0", "orderby=random", "includeallorigins=yes",
	}
	for _, rawQuery := range tests {
		t.Run(rawQuery, func(t *testing.T) {
			response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/event/1/query", rawQuery, ""))
			if response.StatusCode != http.StatusBadRequest {
				t.Fatalf("response = %#v", response)
			}
		})
	}
}

func TestNetHTTPAdapter(t *testing.T) {
	provider := testProvider{station: func(StationQuery) (*stationxmlformat.Document, error) { return stationResult(), nil }}
	server, _ := NewFDSNWS(provider, WithModules(ModuleStation))
	request := httptest.NewRequest(http.MethodGet, "/fdsnws/station/1/query?level=station", nil)
	recorder := httptest.NewRecorder()
	server.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusOK || recorder.Header().Get("Content-Type") != MIMEXML {
		t.Fatalf("recorder = %#v", recorder)
	}
}

func TestNotImplementedAndNil(t *testing.T) {
	server, err := NewFDSNWS(testProvider{}, WithModules(ModuleStation))
	if err != nil {
		t.Fatal(err)
	}
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "", ""))
	if response.StatusCode != http.StatusNotImplemented {
		t.Fatalf("response = %#v", response)
	}
	if _, err = NewFDSNWS(nil); err == nil {
		t.Fatal("expected nil provider error")
	}
	var typedNil *testProvider
	if _, err = NewFDSNWS(typedNil); err == nil {
		t.Fatal("expected typed nil provider error")
	}
	if _, err = NewFDSNWS(testProvider{}, WithServiceVersion(ModuleEvent, "1.x.0")); err == nil {
		t.Fatal("expected invalid version error")
	}
	if _, err = NewFDSNWS(testProvider{}, WithAuthenticator(nil)); err == nil {
		t.Fatal("expected nil authenticator error")
	}
	if !errors.Is(ErrNoData, ErrNoData) {
		t.Fatal("sentinel sanity check")
	}
}

type customService struct{}

func (customService) Name() component.Module  { return "availability" }
func (customService) MajorVersion() string    { return "2" }
func (customService) Version() string         { return "2.0.3" }
func (customService) WADL(base string) string { return "<application base=\"" + base + "\"/>" }
func (customService) Root(versioned bool) (component.Outcome, error) {
	if versioned {
		return component.Outcome{Value: "availability v2", Format: "text"}, nil
	}
	return component.Outcome{Value: "availability", Format: "text"}, nil
}
func (customService) Handle(_ context.Context, endpoint string, _ component.Request) (component.Outcome, error) {
	if endpoint != "query" {
		return component.Outcome{}, &component.RequestError{Status: http.StatusNotFound, Summary: "Not found", Detail: "unknown availability method"}
	}
	return component.Outcome{Value: "available", Format: "text", NoData: http.StatusNoContent}, nil
}

func TestCustomServiceRegistrationDoesNotRequireCoreChanges(t *testing.T) {
	server, err := NewFDSNWS(nil, WithService(customService{}), WithModules("availability"))
	if err != nil {
		t.Fatal(err)
	}
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/availability/2/query", "", ""))
	if response.StatusCode != http.StatusOK || response.MIMEType != MIMEText || string(response.Body) != "available" {
		t.Fatalf("response = %#v", response)
	}
	version := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/availability/2/version", "", ""))
	if string(version.Body) != "2.0.3\n" {
		t.Fatalf("version = %q", version.Body)
	}
}

type stationOnlyProvider struct{}

func (stationOnlyProvider) StationRoot() (component.PreparedResponse, component.ProviderErrorKind, error) {
	return component.PreparedResponse{Body: "station", MIMEType: component.MIMEText, StatusCode: http.StatusOK}, component.ProviderErrorNone, nil
}

func (stationOnlyProvider) StationVersionRoot() (component.PreparedResponse, component.ProviderErrorKind, error) {
	return component.PreparedResponse{Body: "station v1", MIMEType: component.MIMEText, StatusCode: http.StatusOK}, component.ProviderErrorNone, nil
}

func (stationOnlyProvider) Station(component.StationQuery) (*stationxmlformat.Document, component.ProviderErrorKind, error) {
	document := stationxmlformat.NewDocument()
	document.Source = "narrow"
	document.Networks = []stationxmlformat.Network{{Code: "XX"}}
	return document, component.ProviderErrorNone, nil
}

func TestBuiltInComponentAcceptsNarrowProvider(t *testing.T) {
	stationService := stationcomponent.New(stationOnlyProvider{})
	server, err := NewFDSNWS(nil, WithService(stationService), WithModules(ModuleStation))
	if err != nil {
		t.Fatal(err)
	}
	response := server.Handle(context.Background(), normalizedRequest(http.MethodGet, "/fdsnws/station/1/query", "level=network", ""))
	if response.StatusCode != http.StatusOK || response.MIMEType != MIMEXML {
		t.Fatalf("response = %#v", response)
	}
}
