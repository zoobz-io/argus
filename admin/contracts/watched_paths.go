package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// WatchedPaths defines admin operations for watched paths.
type WatchedPaths interface {
	// GetWatchedPath retrieves a watched path by ID.
	GetWatchedPath(ctx context.Context, id string) (*models.WatchedPath, error)
	// CreateWatchedPath creates a new watched path.
	CreateWatchedPath(ctx context.Context, tenantID string, providerID string, path string) (*models.WatchedPath, error)
	// UpdateWatchedPath updates an existing watched path.
	UpdateWatchedPath(ctx context.Context, id string, path string) (*models.WatchedPath, error)
	// DeleteWatchedPath removes a watched path.
	DeleteWatchedPath(ctx context.Context, id string) error
	// ListWatchedPaths retrieves a paginated list of all watched paths.
	ListWatchedPaths(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error)
	// ListWatchedPathsByTenant retrieves watched paths for a specific tenant.
	ListWatchedPathsByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error)
}
