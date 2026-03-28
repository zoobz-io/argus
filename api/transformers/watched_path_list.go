package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// WatchedPathsToListResponse converts an offset result of watched paths to an API list response.
func WatchedPathsToListResponse(result *models.OffsetResult[models.WatchedPath]) wire.WatchedPathListResponse {
	items := make([]wire.WatchedPathResponse, len(result.Items))
	for i, w := range result.Items {
		items[i] = WatchedPathToResponse(w)
	}
	return wire.WatchedPathListResponse{
		WatchedPaths: items,
		Offset:       result.Offset,
		Total:        result.Total,
	}
}
