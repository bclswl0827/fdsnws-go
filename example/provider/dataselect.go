package provider

import (
	"fmt"
	"math"
	"strings"
	"time"

	fdsnws "github.com/bclswl0827/fdsnws-go"
)

const (
	sampleRate       = 100
	samplesPerRecord = 100
)

func (Provider) DataSelect(query fdsnws.DataSelectQuery) (fdsnws.DataSelectResult, fdsnws.ProviderErrorKind, error) {
	if len(query.Selections) == 0 {
		return fdsnws.DataSelectResult{}, fdsnws.ProviderErrorNone, fdsnws.ErrNoData
	}
	result := fdsnws.DataSelectResult{}
	for index, selection := range query.Selections {
		if selection.Start == nil || selection.End == nil {
			return fdsnws.DataSelectResult{}, fdsnws.ProviderErrorClient, fmt.Errorf("selection %d requires starttime and endtime", index+1)
		}
		if (*selection.End).Sub(*selection.Start) > time.Hour {
			return fdsnws.DataSelectResult{}, fdsnws.ProviderErrorClient, fmt.Errorf("time range too long!")
		}
		totalSamples := int(math.Ceil(selection.End.Sub(*selection.Start).Seconds() * sampleRate))
		quality := query.Quality
		if quality == "" || quality == "B" {
			quality = "D"
		}
		for offset := 0; offset < totalSamples; offset += samplesPerRecord {
			count := min(samplesPerRecord, totalSamples-offset)
			recordStart := selection.Start.Add(time.Duration(offset) * time.Second / time.Duration(sampleRate))
			result.Records = append(result.Records, fdsnws.DataRecord{
				Network: seedCode(selection.Network, "XX"), Station: seedCode(selection.Station, "DEMO"),
				Location: seedCode(selection.Location, "00"), Channel: seedCode(selection.Channel, "BHZ"),
				Quality: quality, StartTime: recordStart, SampleRate: sampleRate,
				Samples: sineWave(count, sampleRate, 2, 1000, offset),
			})
		}
	}
	if len(result.Records) == 0 {
		return fdsnws.DataSelectResult{}, fdsnws.ProviderErrorNone, fdsnws.ErrNoData
	}

	return result, fdsnws.ProviderErrorNone, nil
}

func sineWave(length, rate int, frequency, amplitude float64, phaseOffset int) []int32 {
	data := make([]int32, length)
	for index := range data {
		seconds := float64(index+phaseOffset) / float64(rate)
		data[index] = int32(amplitude * math.Sin(2*math.Pi*frequency*seconds))
	}

	return data
}

func seedCode(value, fallback string) string {
	value = strings.TrimSpace(strings.Split(value, ",")[0])
	if value == "" || strings.ContainsAny(value, "*?") {
		return fallback
	}

	return value
}
