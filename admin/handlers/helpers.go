package handlers

import (
	"strconv"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/rocco"
)

// offsetPageFromQuery builds an OffsetPage from query parameters.
func offsetPageFromQuery(params *rocco.Params) models.OffsetPage {
	page := models.OffsetPage{
		Limit: models.DefaultPageSize,
	}
	if v := params.Query["limit"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Limit = n
		}
	}
	if v := params.Query["offset"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Offset = n
		}
	}
	return page
}

// pathID extracts a string path parameter.
func pathID(params *rocco.Params, name string) string {
	return params.Path[name]
}
