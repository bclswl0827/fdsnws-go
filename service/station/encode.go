package station

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bclswl0827/fdsnws-go/service"
	"github.com/bclswl0827/fdsnws-go/stationxml"
)

func encode(query service.StationQuery, document *stationxml.Document, currentTime time.Time) ([]byte, error) {
	if document == nil || len(document.Networks) == 0 {
		return nil, service.ErrNoData
	}
	if query.Format == "text" {
		return encodeText(query.Level, document), nil
	}
	return encodeXML(query, document, currentTime)
}

func encodeText(level string, document *stationxml.Document) []byte {
	var buffer bytes.Buffer
	switch level {
	case "network":
		buffer.WriteString("#Network|Description|StartTime|EndTime|TotalStations\n")
		for _, network := range document.Networks {
			writeFields(&buffer, network.Code, network.Description, network.StartDate, network.EndDate, strconv.Itoa(len(network.Stations)))
		}
	case "channel":
		buffer.WriteString("#Network|Station|Location|Channel|Latitude|Longitude|Elevation|Depth|Azimuth|Dip|SensorDescription|Scale|ScaleFrequency|ScaleUnits|SampleRate|StartTime|EndTime\n")
		for _, network := range document.Networks {
			for _, station := range network.Stations {
				for _, channel := range station.Channels {
					description, scale, frequency, units := "", "", "", ""
					if channel.Sensor != nil {
						description = channel.Sensor.Description
					}
					if channel.Response != nil && channel.Response.Sensitivity != nil {
						sensitivity := channel.Response.Sensitivity
						scale, frequency, units = number(sensitivity.Value), number(sensitivity.Frequency), sensitivity.Input.Name
					}
					writeFields(&buffer, network.Code, station.Code, channel.LocationCode, channel.Code, number(channel.Latitude), number(channel.Longitude), number(channel.Elevation), number(channel.Depth), number(channel.Azimuth), number(channel.Dip), description, scale, frequency, units, number(channel.SampleRate), channel.StartDate, channel.EndDate)
				}
			}
		}
	default:
		buffer.WriteString("#Network|Station|Latitude|Longitude|Elevation|SiteName|StartTime|EndTime\n")
		for _, network := range document.Networks {
			for _, station := range network.Stations {
				writeFields(&buffer, network.Code, station.Code, number(station.Latitude), number(station.Longitude), number(station.Elevation), station.Site.Name, station.StartDate, station.EndDate)
			}
		}
	}
	return buffer.Bytes()
}

func writeFields(buffer *bytes.Buffer, fields ...string) {
	for index, field := range fields {
		if index > 0 {
			buffer.WriteByte('|')
		}
		buffer.WriteString(strings.NewReplacer("|", " ", "\r", " ", "\n", " ").Replace(field))
	}
	buffer.WriteByte('\n')
}

func number(value float64) string { return strconv.FormatFloat(value, 'g', -1, 64) }

func encodeXML(query service.StationQuery, document *stationxml.Document, currentTime time.Time) ([]byte, error) {
	projected := projectDocument(query, document)
	if projected.Source == "" {
		projected.Source = "fdsnws-go"
	}
	if projected.Created == "" {
		projected.Created = currentTime.UTC().Format(time.RFC3339Nano)
	}
	body, err := stationxml.Marshal(projected)
	if err != nil {
		return nil, fmt.Errorf("encode StationXML: %w", err)
	}
	return body, nil
}

func projectDocument(query service.StationQuery, document *stationxml.Document) *stationxml.Document {
	result := *document
	result.Networks = make([]stationxml.Network, 0, len(document.Networks))
	for _, network := range document.Networks {
		networkCopy := network
		networkCopy.Stations = nil
		if query.Level != "network" {
			networkCopy.Stations = make([]stationxml.Station, 0, len(network.Stations))
			for _, station := range network.Stations {
				stationCopy := station
				stationCopy.Channels = nil
				if query.Level == "channel" || query.Level == "response" {
					stationCopy.Channels = make([]stationxml.Channel, 0, len(station.Channels))
					for _, channel := range station.Channels {
						channelCopy := channel
						if !query.IncludeAvailability {
							channelCopy.Availability = nil
						}
						if query.Level != "response" {
							channelCopy.Response = nil
						}
						stationCopy.Channels = append(stationCopy.Channels, channelCopy)
					}
				}
				networkCopy.Stations = append(networkCopy.Stations, stationCopy)
			}
		}
		result.Networks = append(result.Networks, networkCopy)
	}
	return &result
}
