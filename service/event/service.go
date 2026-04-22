package event

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bclswl0827/fdsnws-go/service"
	"github.com/bclswl0827/fdsnws-go/utils"
)

const implementationVersion = "1.2.0"

type Component struct {
	provider service.IEventProvider
}

func New(provider service.IEventProvider) *Component { return &Component{provider: provider} }

func (*Component) Name() service.Module { return service.ModuleEvent }
func (*Component) MajorVersion() string { return "1" }
func (*Component) Version() string      { return implementationVersion }

func (c *Component) Root(versioned bool) (service.Outcome, error) {
	var result service.PreparedResponse
	var kind service.ProviderErrorKind
	var err error
	if versioned {
		result, kind, err = c.provider.EventVersionRoot()
	} else {
		result, kind, err = c.provider.EventRoot()
	}
	return service.Outcome{Value: result.Body, MIMEType: result.MIMEType, StatusCode: result.StatusCode, ErrorKind: kind}, err
}

func (c *Component) Handle(ctx context.Context, endpoint string, request service.Request) (service.Outcome, error) {
	if !strings.EqualFold(request.Method, http.MethodGet) {
		return service.Outcome{}, &service.RequestError{Status: http.StatusMethodNotAllowed, Summary: "Method not allowed", Detail: "Event endpoints accept GET"}
	}
	switch endpoint {
	case "query":
		query, err := parse(request)
		if err != nil {
			return service.Outcome{}, &service.RequestError{Status: http.StatusBadRequest, Summary: "Bad request", Detail: err.Error()}
		}
		if err := ctx.Err(); err != nil {
			return service.Outcome{NoData: query.NoData}, err
		}
		result, kind, err := c.provider.Event(query)
		if err != nil {
			return service.Outcome{NoData: query.NoData, ErrorKind: kind}, err
		}
		value, err := encodeQuery(query, result)
		if err != nil {
			return service.Outcome{NoData: query.NoData}, err
		}
		mimeType := service.MIMEXML
		if query.Format == "text" {
			mimeType = service.MIMEText
		}
		return service.Outcome{Value: value, Format: query.Format, MIMEType: mimeType, NoData: query.NoData}, nil
	case "catalogs":
		if err := ctx.Err(); err != nil {
			return service.Outcome{NoData: http.StatusNoContent}, err
		}
		result, kind, err := c.provider.EventCatalogs()
		if err != nil {
			return service.Outcome{NoData: http.StatusNoContent, ErrorKind: kind}, err
		}
		value, err := encodeCatalogs(result)
		return service.Outcome{Value: value, MIMEType: service.MIMEXML, NoData: http.StatusNoContent}, err
	case "contributors":
		if err := ctx.Err(); err != nil {
			return service.Outcome{NoData: http.StatusNoContent}, err
		}
		result, kind, err := c.provider.EventContributors()
		if err != nil {
			return service.Outcome{NoData: http.StatusNoContent, ErrorKind: kind}, err
		}
		value, err := encodeContributors(result)
		return service.Outcome{Value: value, MIMEType: service.MIMEXML, NoData: http.StatusNoContent}, err
	default:
		return service.Outcome{}, &service.RequestError{Status: http.StatusNotFound, Summary: "Not found", Detail: "Unknown event method"}
	}
}

func (*Component) WADL(basePath string) string {
	parameters := []string{"starttime", "endtime", "minlatitude", "maxlatitude", "minlongitude", "maxlongitude", "latitude", "longitude", "minradius", "maxradius", "mindepth", "maxdepth", "minmagnitude", "maxmagnitude", "magnitudetype", "eventtype", "includeallorigins", "includeallmagnitudes", "includearrivals", "eventid", "limit", "offset", "orderby", "contributor", "catalog", "updatedafter", "format", "nodata"}
	return utils.BuildWADL(basePath+"/event/1/", parameters, false, map[string][]string{"catalogs": {"GET"}, "contributors": {"GET"}})
}

func parse(request service.Request) (service.EventQuery, error) {
	query := service.EventQuery{Offset: 1, OrderBy: "time", Format: "xml", NoData: http.StatusNoContent}
	valid := utils.Allowed("starttime", "endtime", "minlatitude", "maxlatitude", "minlongitude", "maxlongitude", "latitude", "longitude", "minradius", "maxradius", "mindepth", "maxdepth", "minmagnitude", "maxmagnitude", "magnitudetype", "eventtype", "includeallorigins", "includeallmagnitudes", "includearrivals", "eventid", "limit", "offset", "orderby", "contributor", "catalog", "updatedafter", "format", "nodata")
	values, err := utils.RequestValues(request)
	if err != nil {
		return query, err
	}
	params, err := utils.NormalizeParams(values, utils.DefaultAliases(), valid)
	if err != nil {
		return query, err
	}
	if query.TimeRange.Start, err = utils.OptionalTime(params, "starttime"); err != nil {
		return query, err
	}
	if query.TimeRange.End, err = utils.OptionalTime(params, "endtime"); err != nil {
		return query, err
	}
	if err = utils.ValidateTimeRange(query.TimeRange); err != nil {
		return query, err
	}
	if query.Geographic, err = utils.ParseGeographic(params); err != nil {
		return query, err
	}
	if query.MinDepth, err = utils.OptionalFloat(params, "mindepth"); err != nil {
		return query, err
	}
	if query.MaxDepth, err = utils.OptionalFloat(params, "maxdepth"); err != nil {
		return query, err
	}
	if query.MinMagnitude, err = utils.OptionalFloat(params, "minmagnitude"); err != nil {
		return query, err
	}
	if query.MaxMagnitude, err = utils.OptionalFloat(params, "maxmagnitude"); err != nil {
		return query, err
	}
	if err = utils.ValidateBounds("depth", query.MinDepth, query.MaxDepth); err != nil {
		return query, err
	}
	if err = utils.ValidateBounds("magnitude", query.MinMagnitude, query.MaxMagnitude); err != nil {
		return query, err
	}
	query.MagnitudeType = params["magnitudetype"]
	if value := params["eventtype"]; value != "" {
		for _, item := range strings.Split(value, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				return query, fmt.Errorf("eventtype contains an empty item")
			}
			query.EventType = append(query.EventType, item)
		}
	}
	if query.IncludeAllOrigins, err = utils.BoolDefault(params, "includeallorigins", false); err != nil {
		return query, err
	}
	if query.IncludeAllMagnitudes, err = utils.BoolDefault(params, "includeallmagnitudes", false); err != nil {
		return query, err
	}
	if query.IncludeArrivals, err = utils.BoolDefault(params, "includearrivals", false); err != nil {
		return query, err
	}
	query.EventID = params["eventid"]
	if query.Limit, err = utils.OptionalPositiveInt(params, "limit"); err != nil {
		return query, err
	}
	if value, ok := params["offset"]; ok {
		query.Offset, err = utils.ParsePositiveInt("offset", value)
		if err != nil {
			return query, err
		}
	}
	if value, ok := params["orderby"]; ok {
		query.OrderBy = strings.ToLower(value)
	}
	if !utils.OneOf(query.OrderBy, "time", "time-asc", "magnitude", "magnitude-asc") {
		return query, fmt.Errorf("orderby must be time, time-asc, magnitude, or magnitude-asc")
	}
	query.Contributor, query.Catalog = params["contributor"], params["catalog"]
	if query.UpdatedAfter, err = utils.OptionalTime(params, "updatedafter"); err != nil {
		return query, err
	}
	if value, ok := params["format"]; ok {
		query.Format = strings.ToLower(value)
	}
	if !utils.OneOf(query.Format, "xml", "text") {
		return query, fmt.Errorf("format must be xml or text")
	}
	if query.NoData, err = utils.ParseNoData(params); err != nil {
		return query, err
	}
	return query, nil
}
