package stationxml

import "encoding/xml"

const (
	Namespace            = "http://www.fdsn.org/xml/station/1"
	DefaultSchemaVersion = "1.2"
)

type Document struct {
	XMLName       xml.Name  `xml:"FDSNStationXML"`
	XMLNS         string    `xml:"xmlns,attr"`
	SchemaVersion string    `xml:"schemaVersion,attr"`
	Source        string    `xml:"Source"`
	Sender        string    `xml:"Sender,omitempty"`
	Module        string    `xml:"Module,omitempty"`
	ModuleURI     string    `xml:"ModuleURI,omitempty"`
	Created       string    `xml:"Created"`
	Networks      []Network `xml:"Network"`
}

type Network struct {
	Code             string    `xml:"code,attr"`
	StartDate        string    `xml:"startDate,attr,omitempty"`
	EndDate          string    `xml:"endDate,attr,omitempty"`
	RestrictedStatus string    `xml:"restrictedStatus,attr,omitempty"`
	Description      string    `xml:"Description,omitempty"`
	Stations         []Station `xml:"Station,omitempty"`
}

type Station struct {
	Code             string    `xml:"code,attr"`
	StartDate        string    `xml:"startDate,attr,omitempty"`
	EndDate          string    `xml:"endDate,attr,omitempty"`
	RestrictedStatus string    `xml:"restrictedStatus,attr,omitempty"`
	Latitude         float64   `xml:"Latitude"`
	Longitude        float64   `xml:"Longitude"`
	Elevation        float64   `xml:"Elevation"`
	Site             Site      `xml:"Site"`
	Channels         []Channel `xml:"Channel,omitempty"`
}

type Site struct {
	Name string `xml:"Name"`
}

type Channel struct {
	Code             string            `xml:"code,attr"`
	LocationCode     string            `xml:"locationCode,attr"`
	StartDate        string            `xml:"startDate,attr,omitempty"`
	EndDate          string            `xml:"endDate,attr,omitempty"`
	RestrictedStatus string            `xml:"restrictedStatus,attr,omitempty"`
	Latitude         float64           `xml:"Latitude"`
	Longitude        float64           `xml:"Longitude"`
	Elevation        float64           `xml:"Elevation"`
	Depth            float64           `xml:"Depth"`
	Azimuth          float64           `xml:"Azimuth"`
	Dip              float64           `xml:"Dip"`
	SampleRate       float64           `xml:"SampleRate"`
	Sensor           *Sensor           `xml:"Sensor,omitempty"`
	Availability     *DataAvailability `xml:"DataAvailability,omitempty"`
	Response         *Response         `xml:"Response,omitempty"`
}

type Sensor struct {
	Description string `xml:"Description,omitempty"`
}

type DataAvailability struct {
	Extent Extent `xml:"Extent"`
}

type Extent struct {
	Start string `xml:"start,attr"`
	End   string `xml:"end,attr"`
}

type Response struct {
	Sensitivity *InstrumentSensitivity `xml:"InstrumentSensitivity,omitempty"`
	Stages      []Stage                `xml:"Stage,omitempty"`
}

type InstrumentSensitivity struct {
	Value     float64 `xml:"Value"`
	Frequency float64 `xml:"Frequency"`
	Input     Units   `xml:"InputUnits"`
	Output    Units   `xml:"OutputUnits"`
}

type Units struct {
	Name string `xml:"Name"`
}

type Stage struct {
	Number       int           `xml:"number,attr"`
	PolesZeros   *PolesZeros   `xml:"PolesZeros,omitempty"`
	Coefficients *Coefficients `xml:"Coefficients,omitempty"`
	FIR          *FIR          `xml:"FIR,omitempty"`
	Gain         *StageGain    `xml:"StageGain,omitempty"`
}

type PolesZeros struct {
	Input     Units      `xml:"InputUnits"`
	Output    Units      `xml:"OutputUnits"`
	Type      string     `xml:"PzTransferFunctionType"`
	Factor    float64    `xml:"NormalizationFactor"`
	Frequency float64    `xml:"NormalizationFrequency"`
	Zeros     []PoleZero `xml:"Zero,omitempty"`
	Poles     []PoleZero `xml:"Pole,omitempty"`
}

type PoleZero struct {
	Number    int     `xml:"number,attr"`
	Real      float64 `xml:"Real"`
	Imaginary float64 `xml:"Imaginary"`
}

type Coefficients struct {
	Input        Units     `xml:"InputUnits"`
	Output       Units     `xml:"OutputUnits"`
	Type         string    `xml:"CfTransferFunctionType"`
	Numerators   []float64 `xml:"Numerator,omitempty"`
	Denominators []float64 `xml:"Denominator,omitempty"`
}

type FIR struct {
	Symmetry   string    `xml:"symmetry,attr,omitempty"`
	Input      Units     `xml:"InputUnits"`
	Output     Units     `xml:"OutputUnits"`
	Numerators []float64 `xml:"NumeratorCoefficient,omitempty"`
}

type StageGain struct {
	Value     float64 `xml:"Value"`
	Frequency float64 `xml:"Frequency"`
}
