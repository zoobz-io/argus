package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// WatchedPathToResponse transforms a WatchedPath model to an API response.
func WatchedPathToResponse(w *models.WatchedPath) wire.WatchedPathResponse {
	return wire.WatchedPathResponse{
		ID:         w.ID,
		ProviderID: w.ProviderID,
		Path:       w.Path,
		Active:     w.Active,
	}
}

// WatchedPathsToResponse transforms a slice of WatchedPath models to responses.
func WatchedPathsToResponse(paths []*models.WatchedPath) []wire.WatchedPathResponse {
	result := make([]wire.WatchedPathResponse, len(paths))
	for i, w := range paths {
		result[i] = WatchedPathToResponse(w)
	}
	return result
}

// ApplyWatchedPathCreate applies a create request to a WatchedPath model.
func ApplyWatchedPathCreate(req wire.WatchedPathCreateRequest, w *models.WatchedPath) {
	w.ProviderID = req.ProviderID
	w.Path = req.Path
}
