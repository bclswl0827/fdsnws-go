package dataselect

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
	provider      service.IDataSelectProvider
	authenticator service.Authenticator
}

func New(provider service.IDataSelectProvider, authenticators ...service.Authenticator) *Component {
	component := &Component{provider: provider}
	if len(authenticators) > 0 {
		component.authenticator = authenticators[0]
	}
	return component
}

func (*Component) Name() service.Module { return service.ModuleDataSelect }
func (*Component) MajorVersion() string { return "1" }
func (*Component) Version() string      { return implementationVersion }

func (c *Component) Root(versioned bool) (service.Outcome, error) {
	var result service.PreparedResponse
	var kind service.ProviderErrorKind
	var err error
	if versioned {
		result, kind, err = c.provider.DataSelectVersionRoot()
	} else {
		result, kind, err = c.provider.DataSelectRoot()
	}
	return service.Outcome{Value: result.Body, MIMEType: result.MIMEType, StatusCode: result.StatusCode, ErrorKind: kind}, err
}

func (c *Component) Handle(ctx context.Context, endpoint string, request service.Request) (service.Outcome, error) {
	if endpoint == "queryauth" && c.authenticator == nil {
		return service.Outcome{}, &service.RequestError{Status: http.StatusNotFound, Summary: "Not found", Detail: "Dataselect queryauth is not enabled"}
	}
	if endpoint != "query" && endpoint != "queryauth" {
		return service.Outcome{}, &service.RequestError{Status: http.StatusNotFound, Summary: "Not found", Detail: "Unknown dataselect method"}
	}
	if !strings.EqualFold(request.Method, http.MethodGet) && !strings.EqualFold(request.Method, http.MethodPost) {
		return service.Outcome{}, &service.RequestError{Status: http.StatusMethodNotAllowed, Summary: "Method not allowed", Detail: "Dataselect query endpoints accept GET or POST"}
	}
	if err := ctx.Err(); err != nil {
		return service.Outcome{}, err
	}
	var principal *service.Principal
	if endpoint == "queryauth" {
		value, err := c.authenticator.Authenticate(ctx, authenticationRequest(request))
		if err != nil {
			return service.Outcome{}, err
		}
		principal = &value
	}
	query, err := parse(request)
	if err != nil {
		return service.Outcome{}, &service.RequestError{Status: http.StatusBadRequest, Summary: "Bad request", Detail: err.Error()}
	}
	query.Principal = principal
	result, kind, err := c.provider.DataSelect(query)
	if err != nil {
		return service.Outcome{NoData: query.NoData, ErrorKind: kind}, err
	}
	data, err := encodeMiniSEED(query, result)
	if err != nil {
		return service.Outcome{NoData: query.NoData}, err
	}
	return service.Outcome{Value: data, Format: query.Format, MIMEType: service.MIMEMiniSEED, NoData: query.NoData}, nil
}

func (c *Component) WADL(basePath string) string {
	parameters := []string{"starttime", "endtime", "network", "station", "location", "channel", "quality", "minimumlength", "longestonly", "format", "nodata"}
	var extra map[string][]string
	if c.authenticator != nil {
		extra = map[string][]string{"queryauth": {"GET", "POST"}}
	}
	return utils.BuildWADL(basePath+"/dataselect/1/", parameters, true, extra)
}

func authenticationRequest(request service.Request) service.AuthenticationRequest {
	target := request.Path
	if request.RawQuery != "" {
		target += "?" + request.RawQuery
	}
	return service.AuthenticationRequest{
		Method:        request.Method,
		RequestTarget: target,
		Authorization: headerValue(request.Headers, "Authorization"),
		Body:          append([]byte(nil), request.Body...),
		RemoteAddr:    request.RemoteAddr,
	}
}

func headerValue(headers map[string][]string, name string) string {
	for key, values := range headers {
		if strings.EqualFold(key, name) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func parse(request service.Request) (service.DataSelectQuery, error) {
	query := service.DataSelectQuery{Quality: "B", Format: "miniseed", NoData: http.StatusNoContent}
	valid := utils.Allowed("starttime", "endtime", "network", "station", "location", "channel", "quality", "minimumlength", "longestonly", "format", "nodata")
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
	if value, ok := params["quality"]; ok {
		value = strings.ToUpper(value)
		if !utils.OneOf(value, "D", "R", "Q", "M", "B") {
			return query, fmt.Errorf("quality must be one of D, R, Q, M, or B")
		}
		query.Quality = value
	}
	if value, ok := params["minimumlength"]; ok {
		query.MinimumLength, err = utils.ParseFloat("minimumlength", value)
		if err != nil || query.MinimumLength < 0 {
			return query, fmt.Errorf("minimumlength must be a decimal number greater than or equal to 0")
		}
	}
	if value, ok := params["longestonly"]; ok {
		query.LongestOnly, err = utils.ParseBool("longestonly", value)
		if err != nil {
			return query, err
		}
	}
	if value, ok := params["format"]; ok && !strings.EqualFold(value, "miniseed") {
		return query, fmt.Errorf("format must be miniseed")
	}
	if query.NoData, err = utils.ParseNoData(params); err != nil {
		return query, err
	}
	if strings.EqualFold(request.Method, http.MethodPost) {
		if len(rows) == 0 {
			return query, fmt.Errorf("POST body must contain at least one stream selection")
		}
		for index, row := range rows {
			selection, parseErr := utils.ParseSelection(row)
			if parseErr != nil {
				return query, fmt.Errorf("selection %d: %w", index+1, parseErr)
			}
			query.Selections = append(query.Selections, selection)
		}
		return query, nil
	}
	for _, key := range []string{"network", "station", "location", "channel", "starttime", "endtime"} {
		if value, ok := params[key]; !ok || value == "" {
			return query, fmt.Errorf("missing required parameter %q", key)
		}
	}
	selection, err := utils.ParseSelection([]string{params["network"], params["station"], params["location"], params["channel"], params["starttime"], params["endtime"]})
	if err != nil {
		return query, err
	}
	query.Selections = []service.StreamEpoch{selection}
	return query, nil
}
