package service

import "time"

// DataSelectResult contains format-neutral waveform records. The dataselect
// component performs miniSEED encoding.
type DataSelectResult struct {
	Records []DataRecord
}

type DataRecord struct {
	Network    string
	Station    string
	Location   string
	Channel    string
	Quality    string
	StartTime  time.Time
	SampleRate float64
	Samples    []int32
}

type CatalogResult struct {
	Catalogs []string
}

type ContributorResult struct {
	Contributors []string
}
