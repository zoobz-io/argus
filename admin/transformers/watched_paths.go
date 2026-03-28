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

// WatchedPathsToAdminList converts an offset result of watched paths to an admin list response.
func WatchedPathsToAdminList(result *models.OffsetResult[models.WatchedPath]) wire.AdminWatchedPathListResponse {
	items := make([]wire.AdminWatchedPathResponse, len(result.Items))
	for i, w := range result.Items {
		items[i] = WatchedPathToAdminResponse(w)
	}
	return wire.AdminWatchedPathListResponse{
		WatchedPaths: items,
		Offset:       result.Offset,
		Total:        result.Total,
	}
}
