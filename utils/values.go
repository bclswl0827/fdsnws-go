package utils

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bclswl0827/fdsnws-go/service"
)

var decimalPattern = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)$`)

func ParseTime(name, value string) (time.Time, error) {
	layouts := []string{time.RFC3339Nano, "2006-01-02T15:04:05.999999999", "2006-01-02T15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return parsed.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("%s must be an ISO 8601 UTC time", name)
}

func OptionalTime(params Params, name string) (*time.Time, error) {
	value, ok := params[name]
	if !ok || value == "" {
		return nil, nil
	}
	parsed, err := ParseTime(name, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func ParseFloat(name, value string) (float64, error) {
	if !decimalPattern.MatchString(value) {
		return 0, fmt.Errorf("%s must use decimal notation", name)
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid %s: %w", name, err)
	}
	return parsed, nil
}

func OptionalFloat(params Params, name string) (*float64, error) {
	value, ok := params[name]
	if !ok || value == "" {
		return nil, nil
	}
	parsed, err := ParseFloat(name, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func ParseBool(name, value string) (bool, error) {
	switch strings.ToLower(value) {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, fmt.Errorf("%s must be TRUE or FALSE", name)
	}
}

func BoolDefault(params Params, name string, defaultValue bool) (bool, error) {
	value, ok := params[name]
	if !ok {
		return defaultValue, nil
	}
	return ParseBool(name, value)
}

func ParseNoData(params Params) (int, error) {
	value, ok := params["nodata"]
	if !ok || value == "" || value == "204" {
		return http.StatusNoContent, nil
	}
	if value == "404" {
		return http.StatusNotFound, nil
	}
	return 0, fmt.Errorf("nodata must be 204 or 404")
}

func ParsePositiveInt(name, value string) (int, error) {
	if value == "" || strings.ContainsAny(value, ".eE+-") {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return 0, fmt.Errorf("%s must be a positive integer", name)
	}
	return parsed, nil
}

func OptionalPositiveInt(params Params, name string) (*int, error) {
	value, ok := params[name]
	if !ok {
		return nil, nil
	}
	parsed, err := ParsePositiveInt(name, value)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func ValidateTimeRange(value service.TimeRange) error {
	if value.Start != nil && value.End != nil && value.Start.After(*value.End) {
		return fmt.Errorf("starttime must not be after endtime")
	}
	return nil
}

func ValidateBounds(name string, minimum, maximum *float64) error {
	if minimum != nil && maximum != nil && *minimum > *maximum {
		return fmt.Errorf("minimum %s must not exceed maximum %s", name, name)
	}
	return nil
}

func NormalizeLocation(value string) string {
	if value == "--" {
		return "  "
	}
	return value
}

func ValidateSelector(name, value string) error {
	for _, char := range value {
		if char < 32 || char > 126 {
			return fmt.Errorf("%s must contain only printable ASCII characters", name)
		}
	}
	return nil
}

func OneOf(value string, allowed ...string) bool {
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}
