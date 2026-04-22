package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/bclswl0827/fdsnws-go/service"
)

type Params map[string]string

var defaultAliases = map[string]string{
	"start": "starttime", "end": "endtime",
	"net": "network", "sta": "station", "loc": "location", "cha": "channel",
	"minlat": "minlatitude", "maxlat": "maxlatitude", "minlon": "minlongitude", "maxlon": "maxlongitude",
	"lat": "latitude", "lon": "longitude",
	"minmag": "minmagnitude", "maxmag": "maxmagnitude", "magtype": "magnitudetype",
}

func DefaultAliases() map[string]string {
	result := make(map[string]string, len(defaultAliases))
	for key, value := range defaultAliases {
		result[key] = value
	}
	return result
}

func Allowed(keys ...string) map[string]bool {
	result := make(map[string]bool, len(keys))
	for _, key := range keys {
		result[key] = true
	}
	return result
}

func RequestValues(request service.Request) (url.Values, error) {
	values := request.Query
	if values == nil {
		var err error
		values, err = url.ParseQuery(request.RawQuery)
		if err != nil {
			return nil, fmt.Errorf("invalid query string: %w", err)
		}
	}
	copyValues := make(url.Values, len(values))
	for key, entries := range values {
		copyValues[key] = append([]string(nil), entries...)
	}
	return copyValues, nil
}

func NormalizeParams(values url.Values, aliases map[string]string, allowed map[string]bool) (Params, error) {
	params := make(Params, len(values))
	for rawKey, values := range values {
		key := strings.ToLower(rawKey)
		if alias, ok := aliases[key]; ok {
			key = alias
		}
		if !allowed[key] {
			return nil, fmt.Errorf("unrecognized parameter %q", rawKey)
		}
		if len(values) != 1 {
			return nil, fmt.Errorf("parameter %q must be specified exactly once", rawKey)
		}
		if _, exists := params[key]; exists {
			return nil, fmt.Errorf("parameter %q was specified using more than one name", key)
		}
		params[key] = values[0]
	}
	return params, nil
}

func ParsePOSTBody(body []byte, aliases map[string]string, allowed map[string]bool) (Params, [][]string, error) {
	params := make(Params)
	var selections [][]string
	scanner := bufio.NewScanner(bytes.NewReader(body))
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for lineNumber := 1; scanner.Scan(); lineNumber++ {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			rawKey := strings.TrimSpace(parts[0])
			key := strings.ToLower(rawKey)
			if alias, ok := aliases[key]; ok {
				key = alias
			}
			if !allowed[key] {
				return nil, nil, fmt.Errorf("line %d: unrecognized parameter %q", lineNumber, rawKey)
			}
			if _, exists := params[key]; exists {
				return nil, nil, fmt.Errorf("line %d: parameter %q was specified more than once", lineNumber, key)
			}
			params[key] = strings.TrimSpace(parts[1])
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 6 {
			return nil, nil, fmt.Errorf("line %d: selection must contain NET STA LOC CHA STARTTIME ENDTIME", lineNumber)
		}
		selections = append(selections, fields)
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, fmt.Errorf("read POST body: %w", err)
	}
	return params, selections, nil
}
