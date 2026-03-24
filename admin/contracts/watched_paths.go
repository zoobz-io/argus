package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// WatchedPaths defines admin operations for watched paths.
type WatchedPaths interface {
	// GetWatchedPath retrieves a watched path by ID.
	GetWatchedPath(ctx context.Context, id int64) (*models.WatchedPath, error)
	// CreateWatchedPath creates a new watched path.
	CreateWatchedPath(ctx context.Context, tenantID int64, providerID int64, path string) (*models.WatchedPath, error)
	// UpdateWatchedPath updates an existing watched path.
	UpdateWatchedPath(ctx context.Context, id int64, path string) (*models.WatchedPath, error)
	// DeleteWatchedPath removes a watched path.
	DeleteWatchedPath(ctx context.Context, id int64) error
	// ListWatchedPaths retrieves a paginated list of all watched paths.
	ListWatchedPaths(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.WatchedPath], error)
	// ListWatchedPathsByTenant retrieves watched paths for a specific tenant.
	ListWatchedPathsByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.WatchedPath], error)
}
