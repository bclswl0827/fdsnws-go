package event

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"

	"github.com/bclswl0827/fdsnws-go/quakeml"
	"github.com/bclswl0827/fdsnws-go/service"
)

func encodeQuery(query service.EventQuery, document *quakeml.Document) ([]byte, error) {
	if document == nil || len(document.Parameters.Events) == 0 {
		return nil, service.ErrNoData
	}
	if query.Format == "text" {
		return encodeText(document), nil
	}
	return encodeQuakeML(query, document)
}

func encodeText(document *quakeml.Document) []byte {
	var buffer bytes.Buffer
	buffer.WriteString("#EventID|Time|Latitude|Longitude|Depth/km|Author|Catalog|Contributor|ContributorID|MagType|Magnitude|MagAuthor|EventLocationName|EventType\n")
	for _, event := range document.Parameters.Events {
		origin := preferredOrigin(event)
		magnitude := preferredMagnitude(event)
		description, contributor := "", ""
		if event.Description != nil {
			description = event.Description.Text
		}
		if event.CreationInfo != nil {
			contributor = event.CreationInfo.AgencyID
		}
		fields := []string{event.PublicID, "", "", "", "", "", "", contributor, event.PublicID, "", "", "", description, event.Type}
		if origin != nil {
			author := ""
			if origin.CreationInfo != nil {
				author = origin.CreationInfo.Author
			}
			fields[1] = origin.Time.Value
			fields[2], fields[3], fields[4], fields[5] = number(origin.Latitude.Value), number(origin.Longitude.Value), number(origin.Depth.Value/1000), author
		}
		if magnitude != nil {
			author := ""
			if magnitude.CreationInfo != nil {
				author = magnitude.CreationInfo.Author
			}
			fields[9], fields[10], fields[11] = magnitude.Type, number(magnitude.Mag.Value), author
		}
		writeFields(&buffer, fields...)
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

func encodeQuakeML(query service.EventQuery, document *quakeml.Document) ([]byte, error) {
	body, err := quakeml.Marshal(projectDocument(query, document))
	if err != nil {
		return nil, fmt.Errorf("encode QuakeML: %w", err)
	}
	return body, nil
}

func projectDocument(query service.EventQuery, document *quakeml.Document) *quakeml.Document {
	result := *document
	result.Parameters = document.Parameters
	result.Parameters.Events = make([]quakeml.Event, 0, len(document.Parameters.Events))
	for eventIndex, event := range document.Parameters.Events {
		eventCopy := event
		if eventCopy.PublicID == "" {
			eventCopy.PublicID = fmt.Sprintf("smi:local/event/%d", eventIndex+1)
		}
		eventCopy.Origins = selectedOrigins(query, event)
		for index := range eventCopy.Origins {
			if !query.IncludeArrivals {
				eventCopy.Origins[index].Arrivals = nil
			}
		}
		eventCopy.Magnitudes = selectedMagnitudes(query, event)
		result.Parameters.Events = append(result.Parameters.Events, eventCopy)
	}
	return &result
}

func selectedOrigins(query service.EventQuery, event quakeml.Event) []quakeml.Origin {
	if query.IncludeAllOrigins {
		return append([]quakeml.Origin(nil), event.Origins...)
	}
	if value := preferredOrigin(event); value != nil {
		return []quakeml.Origin{*value}
	}
	return nil
}

func selectedMagnitudes(query service.EventQuery, event quakeml.Event) []quakeml.Magnitude {
	if query.IncludeAllMagnitudes {
		return append([]quakeml.Magnitude(nil), event.Magnitudes...)
	}
	if value := preferredMagnitude(event); value != nil {
		return []quakeml.Magnitude{*value}
	}
	return nil
}

func preferredOrigin(event quakeml.Event) *quakeml.Origin {
	for index := range event.Origins {
		if event.Origins[index].PublicID == event.PreferredOriginID {
			return &event.Origins[index]
		}
	}
	if len(event.Origins) > 0 {
		return &event.Origins[0]
	}
	return nil
}

func preferredMagnitude(event quakeml.Event) *quakeml.Magnitude {
	for index := range event.Magnitudes {
		if event.Magnitudes[index].PublicID == event.PreferredMagnitudeID {
			return &event.Magnitudes[index]
		}
	}
	if len(event.Magnitudes) > 0 {
		return &event.Magnitudes[0]
	}
	return nil
}

func encodeCatalogs(result service.CatalogResult) ([]byte, error) {
	if len(result.Catalogs) == 0 {
		return nil, service.ErrNoData
	}
	return encodeList("Catalogs", "Catalog", result.Catalogs)
}

func encodeContributors(result service.ContributorResult) ([]byte, error) {
	if len(result.Contributors) == 0 {
		return nil, service.ErrNoData
	}
	return encodeList("Contributors", "Contributor", result.Contributors)
}

func encodeList(root, item string, values []string) ([]byte, error) {
	var buffer bytes.Buffer
	buffer.WriteString(xml.Header)
	encoder := xml.NewEncoder(&buffer)
	encoder.Indent("", "  ")
	start := xml.StartElement{Name: xml.Name{Local: root}}
	if err := encoder.EncodeToken(start); err != nil {
		return nil, err
	}
	for _, value := range values {
		if err := encoder.EncodeElement(value, xml.StartElement{Name: xml.Name{Local: item}}); err != nil {
			return nil, err
		}
	}
	if err := encoder.EncodeToken(start.End()); err != nil {
		return nil, err
	}
	if err := encoder.Flush(); err != nil {
		return nil, err
	}
	buffer.WriteByte('\n')
	return buffer.Bytes(), nil
}
