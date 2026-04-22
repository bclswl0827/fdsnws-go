package quakeml

import "encoding/xml"

const (
	Namespace    = "http://quakeml.org/xmlns/quakeml/1.2"
	BEDNamespace = "http://quakeml.org/xmlns/bed/1.2"
)

type Document struct {
	XMLName    xml.Name        `xml:"quakeml"`
	XMLNS      string          `xml:"xmlns,attr"`
	Parameters EventParameters `xml:"eventParameters"`
}

type EventParameters struct {
	XMLName  xml.Name `xml:"eventParameters"`
	XMLNS    string   `xml:"xmlns,attr"`
	PublicID string   `xml:"publicID,attr"`
	Events   []Event  `xml:"event"`
}

type Event struct {
	PublicID             string        `xml:"publicID,attr"`
	PreferredOriginID    string        `xml:"preferredOriginID,omitempty"`
	PreferredMagnitudeID string        `xml:"preferredMagnitudeID,omitempty"`
	Type                 string        `xml:"type,omitempty"`
	Description          *Description  `xml:"description,omitempty"`
	Origins              []Origin      `xml:"origin,omitempty"`
	Magnitudes           []Magnitude   `xml:"magnitude,omitempty"`
	CreationInfo         *CreationInfo `xml:"creationInfo,omitempty"`
}

type Description struct {
	Text string `xml:"text"`
	Type string `xml:"type,omitempty"`
}

type Origin struct {
	PublicID     string        `xml:"publicID,attr"`
	Time         TimeQuantity  `xml:"time"`
	Latitude     RealQuantity  `xml:"latitude"`
	Longitude    RealQuantity  `xml:"longitude"`
	Depth        RealQuantity  `xml:"depth"`
	Arrivals     []Arrival     `xml:"arrival,omitempty"`
	CreationInfo *CreationInfo `xml:"creationInfo,omitempty"`
}

type Magnitude struct {
	PublicID     string        `xml:"publicID,attr"`
	Mag          RealQuantity  `xml:"mag"`
	Type         string        `xml:"type,omitempty"`
	OriginID     string        `xml:"originID,omitempty"`
	CreationInfo *CreationInfo `xml:"creationInfo,omitempty"`
}

type Arrival struct {
	PickID       string   `xml:"pickID"`
	Phase        string   `xml:"phase"`
	Azimuth      *float64 `xml:"azimuth,omitempty"`
	Distance     *float64 `xml:"distance,omitempty"`
	TimeResidual *float64 `xml:"timeResidual,omitempty"`
}

type TimeQuantity struct {
	Value string `xml:"value"`
}

type RealQuantity struct {
	Value float64 `xml:"value"`
}

type CreationInfo struct {
	AgencyID string `xml:"agencyID,omitempty"`
	Author   string `xml:"author,omitempty"`
}
