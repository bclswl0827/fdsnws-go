package provider

import (
	"time"

	fdsnws "github.com/bclswl0827/fdsnws-go"
	"github.com/bclswl0827/fdsnws-go/stationxml"
)

func (provider Provider) Station(query fdsnws.StationQuery) (*stationxml.Document, fdsnws.ProviderErrorKind, error) {
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339)
	document := stationxml.NewDocument()
	document.Source = "fdsnws-go example"
	document.Created = provider.Now().Format(time.RFC3339Nano)
	document.Networks = []stationxml.Network{{
		Code: "XX", Description: "Example network", StartDate: start,
		Stations: []stationxml.Station{{
			Code: "DEMO", Latitude: 25.033, Longitude: 121.565,
			Elevation: 10, Site: stationxml.Site{Name: "Example station"}, StartDate: start,
			Channels: []stationxml.Channel{{
				Code: "BHZ", LocationCode: "00", Latitude: 25.033,
				Longitude: 121.565, Elevation: 10, Depth: 0,
				Azimuth: 0, Dip: -90, SampleRate: 100,
				Sensor: &stationxml.Sensor{Description: "Example sensor"}, StartDate: start,
				Response: &stationxml.Response{
					Sensitivity: &stationxml.InstrumentSensitivity{
						Value: 1, Frequency: 1,
						Input: stationxml.Units{Name: "m/s"}, Output: stationxml.Units{Name: "COUNTS"},
					},
					Stages: []stationxml.Stage{{
						Number: 1,
						Gain:   &stationxml.StageGain{Value: 1, Frequency: 1},
					}},
				},
			}},
		}},
	}}

	return document, fdsnws.ProviderErrorNone, nil
}
