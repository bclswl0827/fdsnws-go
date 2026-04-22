package station

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bclswl0827/fdsnws-go/service"
	"github.com/bclswl0827/fdsnws-go/utils"
)

const implementationVersion = "1.1.0"

type Component struct {
	provider service.IStationProvider
}

func New(provider service.IStationProvider) *Component { return &Component{provider: provider} }

func (*Component) Name() service.Module { return service.ModuleStation }
func (*Component) MajorVersion() string { return "1" }
func (*Component) Version() string      { return implementationVersion }

func (c *Component) Root(versioned bool) (service.Outcome, error) {
	var result service.PreparedResponse
	var kind service.ProviderErrorKind
	var err error
	if versioned {
		result, kind, err = c.provider.StationVersionRoot()
	} else {
		result, kind, err = c.provider.StationRoot()
	}
	return service.Outcome{Value: result.Body, MIMEType: result.MIMEType, StatusCode: result.StatusCode, ErrorKind: kind}, err
}

func (c *Component) Handle(ctx context.Context, endpoint string, request service.Request) (service.Outcome, error) {
	if endpoint != "query" {
		return service.Outcome{}, &service.RequestError{Status: http.StatusNotFound, Summary: "Not found", Detail: "Unknown station method"}
	}
	if !strings.EqualFold(request.Method, http.MethodGet) && !strings.EqualFold(request.Method, http.MethodPost) {
		return service.Outcome{}, &service.RequestError{Status: http.StatusMethodNotAllowed, Summary: "Method not allowed", Detail: "Station query accepts GET or POST"}
	}
	query, err := parse(request)
	if err != nil {
		return service.Outcome{}, &service.RequestError{Status: http.StatusBadRequest, Summary: "Bad request", Detail: err.Error()}
	}
	if err := ctx.Err(); err != nil {
		return service.Outcome{NoData: query.NoData}, err
	}
	result, kind, err := c.provider.Station(query)
	if err != nil {
		return service.Outcome{NoData: query.NoData, ErrorKind: kind}, err
	}
	value, err := encode(query, result, request.ReceivedAt)
	if err != nil {
		return service.Outcome{NoData: query.NoData}, err
	}
	mimeType := service.MIMEXML
	if query.Format == "text" {
		mimeType = service.MIMEText
	}
	return service.Outcome{Value: value, Format: query.Format, MIMEType: mimeType, NoData: query.NoData}, nil
}

func (*Component) WADL(basePath string) string {
	parameters := []string{"starttime", "endtime", "startbefore", "startafter", "endbefore", "endafter", "network", "station", "location", "channel", "minlatitude", "maxlatitude", "minlongitude", "maxlongitude", "latitude", "longitude", "minradius", "maxradius", "level", "includerestricted", "includeavailability", "updatedafter", "matchtimeseries", "format", "nodata"}
	return utils.BuildWADL(basePath+"/station/1/", parameters, true, nil)
}

func parse(request service.Request) (service.StationQuery, error) {
	query := service.StationQuery{Level: "station", IncludeRestricted: true, Format: "xml", NoData: http.StatusNoContent}
	valid := utils.Allowed("starttime", "endtime", "startbefore", "startafter", "endbefore", "endafter", "network", "station", "location", "channel", "minlatitude", "maxlatitude", "minlongitude", "maxlongitude", "latitude", "longitude", "minradius", "maxradius", "level", "includerestricted", "includeavailability", "updatedafter", "matchtimeseries", "format", "nodata")
	var params utils.Params
	var rows [][]string
	var err error
	if strings.EqualFold(request.Method, http.MethodPost) {
		params, rows, err = utils.ParsePOSTBody(request.Body, utils.DefaultAliases(), valid)
	} else {
		values, valuesErr := utils.RequestValues(request)
		err = valuesErr
		if err == nil {
			params, err = utils.NormalizeParams(values, utils.DefaultAliases(), valid)
		}
	}
	if err != nil {
		return query, err
	}
	if value, ok := params["level"]; ok {
		query.Level = strings.ToLower(value)
	}
	if !utils.OneOf(query.Level, "network", "station", "channel", "response") {
		return query, fmt.Errorf("level must be network, station, channel, or response")
	}
	if value, ok := params["format"]; ok {
		query.Format = strings.ToLower(value)
	}
	if !utils.OneOf(query.Format, "xml", "text") {
		return query, fmt.Errorf("format must be xml or text")
	}
	if query.Format == "text" && query.Level == "response" {
		return query, fmt.Errorf("format=text is not available for level=response")
	}
	if query.IncludeRestricted, err = utils.BoolDefault(params, "includerestricted", true); err != nil {
		return query, err
	}
	if query.IncludeAvailability, err = utils.BoolDefault(params, "includeavailability", false); err != nil {
		return query, err
	}
	if query.MatchTimeSeries, err = utils.BoolDefault(params, "matchtimeseries", false); err != nil {
		return query, err
	}
	if query.NoData, err = utils.ParseNoData(params); err != nil {
		return query, err
	}
	if query.StartBefore, err = utils.OptionalTime(params, "startbefore"); err != nil {
		return query, err
	}
	if query.StartAfter, err = utils.OptionalTime(params, "startafter"); err != nil {
		return query, err
	}
	if query.EndBefore, err = utils.OptionalTime(params, "endbefore"); err != nil {
		return query, err
	}
	if query.EndAfter, err = utils.OptionalTime(params, "endafter"); err != nil {
		return query, err
	}
	if query.UpdatedAfter, err = utils.OptionalTime(params, "updatedafter"); err != nil {
		return query, err
	}
	simpleTime := len(rows) > 0 || params["starttime"] != "" || params["endtime"] != ""
	windowTime := query.StartBefore != nil || query.StartAfter != nil || query.EndBefore != nil || query.EndAfter != nil
	if simpleTime && windowTime {
		return query, fmt.Errorf("simple-time and window-time constraints are mutually exclusive")
	}
	if query.Geographic, err = utils.ParseGeographic(params); err != nil {
		return query, err
	}
	if len(rows) > 0 {
		for index, row := range rows {
			selection, parseErr := utils.ParseSelection(row)
			if parseErr != nil {
				return query, fmt.Errorf("selection %d: %w", index+1, parseErr)
			}
			query.Selections = append(query.Selections, selection)
		}
	} else {
		selection := service.StreamEpoch{Network: params["network"], Station: params["station"], Location: utils.NormalizeLocation(params["location"]), Channel: params["channel"]}
		for name, value := range map[string]string{"network": selection.Network, "station": selection.Station, "location": selection.Location, "channel": selection.Channel} {
			if value != "" {
				if err = utils.ValidateSelector(name, value); err != nil {
					return query, err
				}
			}
		}
		selection.Start, err = utils.OptionalTime(params, "starttime")
		if err != nil {
			return query, err
		}
		selection.End, err = utils.OptionalTime(params, "endtime")
		if err != nil {
			return query, err
		}
		if err = utils.ValidateTimeRange(selection.TimeRange); err != nil {
			return query, err
		}
		query.Selections = []service.StreamEpoch{selection}
	}
	return query, nil
}
