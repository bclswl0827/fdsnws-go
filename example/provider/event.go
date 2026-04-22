package provider

import (
	"time"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/quakeml"
)

func (Provider) Event(query fdsnws.EventQuery) (*quakeml.Document, fdsnws.ProviderErrorKind, error) {
	originTime := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	document := quakeml.NewDocument("smi:local/fdsnws-go/example")
	document.Parameters.Events = []quakeml.Event{{
		PublicID: "smi:local/example-1", Type: "earthquake",
		Description:       &quakeml.Description{Text: "Taipei", Type: "earthquake name"},
		PreferredOriginID: "smi:local/origin-1", PreferredMagnitudeID: "smi:local/magnitude-1",
		CreationInfo: &quakeml.CreationInfo{AgencyID: "fdsnws-go"},
		Origins: []quakeml.Origin{{
			PublicID: "smi:local/origin-1",
			Time:     quakeml.TimeQuantity{Value: originTime.Format(time.RFC3339Nano)},
			Latitude: quakeml.RealQuantity{Value: 25.033}, Longitude: quakeml.RealQuantity{Value: 121.565},
			Depth: quakeml.RealQuantity{Value: 10_000}, CreationInfo: &quakeml.CreationInfo{Author: "example"},
		}},
		Magnitudes: []quakeml.Magnitude{{
			PublicID: "smi:local/magnitude-1", OriginID: "smi:local/origin-1",
			Type: "Mw", Mag: quakeml.RealQuantity{Value: 5}, CreationInfo: &quakeml.CreationInfo{Author: "example"},
		}},
	}}

	return document, fdsnws.ProviderErrorNone, nil
}

func (Provider) EventCatalogs() (fdsnws.CatalogResult, fdsnws.ProviderErrorKind, error) {
	return fdsnws.CatalogResult{Catalogs: []string{"example"}}, fdsnws.ProviderErrorNone, nil
}

func (Provider) EventContributors() (fdsnws.ContributorResult, fdsnws.ProviderErrorKind, error) {
	return fdsnws.ContributorResult{Contributors: []string{"fdsnws-go"}}, fdsnws.ProviderErrorNone, nil
}
