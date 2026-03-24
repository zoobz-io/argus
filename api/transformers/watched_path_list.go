package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// WatchedPathsToListResponse converts a cursor result of watched paths to an API list response.
func WatchedPathsToListResponse(result *models.CursorResult[models.WatchedPath], limit int) wire.WatchedPathListResponse {
	items := make([]wire.WatchedPathResponse, len(result.Items))
	for i, w := range result.Items {
		items[i] = WatchedPathToResponse(w)
	}
	return wire.WatchedPathListResponse{
		WatchedPaths: items,
		Cursor:       result.Cursor,
		Limit:        limit,
		HasMore:      result.HasMore,
	}
}
