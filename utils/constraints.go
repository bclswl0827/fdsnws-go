package utils

import (
	"fmt"

	"github.com/bclswl0827/fdsnws-go/service"
)

func ParseSelection(fields []string) (service.StreamEpoch, error) {
	if len(fields) != 6 {
		return service.StreamEpoch{}, fmt.Errorf("selection must contain NET STA LOC CHA STARTTIME ENDTIME")
	}
	selection := service.StreamEpoch{Network: fields[0], Station: fields[1], Location: NormalizeLocation(fields[2]), Channel: fields[3]}
	if selection.Network == "" || selection.Station == "" || selection.Location == "" || selection.Channel == "" {
		return selection, fmt.Errorf("network, station, location, and channel cannot be empty")
	}
	for name, value := range map[string]string{"network": selection.Network, "station": selection.Station, "location": selection.Location, "channel": selection.Channel} {
		if err := ValidateSelector(name, value); err != nil {
			return selection, err
		}
	}
	start, err := ParseTime("starttime", fields[4])
	if err != nil {
		return selection, err
	}
	end, err := ParseTime("endtime", fields[5])
	if err != nil {
		return selection, err
	}
	selection.Start, selection.End = &start, &end
	if err := ValidateTimeRange(selection.TimeRange); err != nil {
		return selection, err
	}
	return selection, nil
}

func ParseGeographic(params Params) (service.GeographicConstraints, error) {
	var geo service.GeographicConstraints
	var err error
	if geo.MinLatitude, err = OptionalFloat(params, "minlatitude"); err != nil {
		return geo, err
	}
	if geo.MaxLatitude, err = OptionalFloat(params, "maxlatitude"); err != nil {
		return geo, err
	}
	if geo.MinLongitude, err = OptionalFloat(params, "minlongitude"); err != nil {
		return geo, err
	}
	if geo.MaxLongitude, err = OptionalFloat(params, "maxlongitude"); err != nil {
		return geo, err
	}
	if geo.Latitude, err = OptionalFloat(params, "latitude"); err != nil {
		return geo, err
	}
	if geo.Longitude, err = OptionalFloat(params, "longitude"); err != nil {
		return geo, err
	}
	if geo.MinRadius, err = OptionalFloat(params, "minradius"); err != nil {
		return geo, err
	}
	if geo.MaxRadius, err = OptionalFloat(params, "maxradius"); err != nil {
		return geo, err
	}
	for name, value := range map[string]*float64{"minlatitude": geo.MinLatitude, "maxlatitude": geo.MaxLatitude, "latitude": geo.Latitude} {
		if value != nil && (*value < -90 || *value > 90) {
			return geo, fmt.Errorf("%s must be between -90 and 90", name)
		}
	}
	for name, value := range map[string]*float64{"minlongitude": geo.MinLongitude, "maxlongitude": geo.MaxLongitude, "longitude": geo.Longitude} {
		if value != nil && (*value < -180 || *value > 180) {
			return geo, fmt.Errorf("%s must be between -180 and 180", name)
		}
	}
	if err = ValidateBounds("latitude", geo.MinLatitude, geo.MaxLatitude); err != nil {
		return geo, err
	}
	if err = ValidateBounds("longitude", geo.MinLongitude, geo.MaxLongitude); err != nil {
		return geo, err
	}
	rectangle := geo.MinLatitude != nil || geo.MaxLatitude != nil || geo.MinLongitude != nil || geo.MaxLongitude != nil
	circle := geo.Latitude != nil || geo.Longitude != nil || geo.MinRadius != nil || geo.MaxRadius != nil
	if rectangle && circle {
		return geo, fmt.Errorf("rectangle and circle geographic constraints are mutually exclusive")
	}
	if circle && (geo.Latitude == nil || geo.Longitude == nil) {
		return geo, fmt.Errorf("latitude and longitude are both required for a radius search")
	}
	for name, value := range map[string]*float64{"minradius": geo.MinRadius, "maxradius": geo.MaxRadius} {
		if value != nil && (*value < 0 || *value > 180) {
			return geo, fmt.Errorf("%s must be between 0 and 180", name)
		}
	}
	if err = ValidateBounds("radius", geo.MinRadius, geo.MaxRadius); err != nil {
		return geo, err
	}
	return geo, nil
}
