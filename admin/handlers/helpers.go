package handlers

import (
	"strconv"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/rocco"
)

// cursorPageFromQuery builds a CursorPage from query parameters.
func cursorPageFromQuery(params *rocco.Params) models.CursorPage {
	page := models.CursorPage{
		Limit: models.DefaultPageSize,
	}
	if v := params.Query["limit"]; v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			page.Limit = n
		}
	}
	if v := params.Query["cursor"]; v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			page.Cursor = &n
		}
	}
	return page
}

// pathID parses an int64 path parameter.
func pathID(params *rocco.Params, name string) (int64, error) {
	return strconv.ParseInt(params.Path[name], 10, 64)
}
