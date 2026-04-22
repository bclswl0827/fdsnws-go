package dataselect

import (
	"fmt"
	"strings"

	"github.com/bclswl0827/fdsnws-go/service"
	"github.com/bclswl0827/mseedio"
)

func encodeMiniSEED(query service.DataSelectQuery, result service.DataSelectResult) ([]byte, error) {
	records := filterRecords(query, result.Records)
	if len(records) == 0 {
		return nil, service.ErrNoData
	}
	var miniSeed mseedio.MiniSeedData
	if err := miniSeed.Init(mseedio.INT32, mseedio.MSBFIRST); err != nil {
		return nil, fmt.Errorf("initialize miniSEED: %w", err)
	}
	for index, record := range records {
		if record.SampleRate <= 0 {
			return nil, fmt.Errorf("record %d has an invalid sample rate", index+1)
		}
		if len(record.Samples) == 0 {
			continue
		}
		if record.StartTime.IsZero() {
			return nil, fmt.Errorf("record %d has no start time", index+1)
		}
		err := miniSeed.Append(record.Samples, &mseedio.AppendOptions{
			SampleRate: record.SampleRate, StartTime: record.StartTime,
			SequenceNumber: fmt.Sprintf("%06d", index+1),
			NetworkCode:    seedField(record.Network, 2), StationCode: seedField(record.Station, 5),
			LocationCode: seedField(record.Location, 2), ChannelCode: seedField(record.Channel, 3),
		})
		if err != nil {
			return nil, fmt.Errorf("append miniSEED record %d: %w", index+1, err)
		}
		miniSeed.Series[len(miniSeed.Series)-1].FixedSection.DataQuality = recordQuality(record)
	}
	if miniSeed.Records == 0 {
		return nil, service.ErrNoData
	}
	data, err := miniSeed.Encode(mseedio.OVERWRITE, mseedio.MSBFIRST)
	if err != nil {
		return nil, fmt.Errorf("encode miniSEED: %w", err)
	}
	return data, nil
}

func filterRecords(query service.DataSelectQuery, records []service.DataRecord) []service.DataRecord {
	filtered := make([]service.DataRecord, 0, len(records))
	for _, record := range records {
		quality := recordQuality(record)
		if query.Quality != "B" && quality != query.Quality {
			continue
		}
		if query.MinimumLength > 0 && record.SampleRate > 0 && float64(len(record.Samples))/record.SampleRate < query.MinimumLength {
			continue
		}
		filtered = append(filtered, record)
	}
	if !query.LongestOnly {
		return filtered
	}
	longest := make(map[string]service.DataRecord)
	order := make([]string, 0)
	for _, record := range filtered {
		key := record.Network + "\x00" + record.Station + "\x00" + record.Location + "\x00" + record.Channel
		current, exists := longest[key]
		if !exists {
			order = append(order, key)
		}
		if !exists || duration(record) > duration(current) {
			longest[key] = record
		}
	}
	filtered = filtered[:0]
	for _, key := range order {
		filtered = append(filtered, longest[key])
	}
	return filtered
}

func recordQuality(record service.DataRecord) string {
	quality := strings.ToUpper(strings.TrimSpace(record.Quality))
	if quality == "D" || quality == "R" || quality == "Q" || quality == "M" {
		return quality
	}
	return "D"
}

func duration(record service.DataRecord) float64 {
	if record.SampleRate <= 0 {
		return 0
	}
	return float64(len(record.Samples)) / record.SampleRate
}

func seedField(value string, width int) string {
	value = strings.ToUpper(strings.TrimSpace(value))
	if len(value) > width {
		value = value[:width]
	}
	return value
}
