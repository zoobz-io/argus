package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// WatchedPathToAdminResponse converts a watched path model to an admin API response.
func WatchedPathToAdminResponse(w *models.WatchedPath) wire.AdminWatchedPathResponse {
	return wire.AdminWatchedPathResponse{
		ID:         w.ID,
		TenantID:   w.TenantID,
		ProviderID: w.ProviderID,
		Path:       w.Path,
		Active:     w.Active,
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

// WatchedPathsToAdminList converts a slice of watched path models to a paginated admin list response.
func WatchedPathsToAdminList(paths []*models.WatchedPath, total, limit, offset int) wire.AdminWatchedPathListResponse {
	items := make([]wire.AdminWatchedPathResponse, len(paths))
	for i, w := range paths {
		items[i] = WatchedPathToAdminResponse(w)
	}
	return wire.AdminWatchedPathListResponse{
		WatchedPaths: items,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}
}
