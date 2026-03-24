package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// WatchedPaths defines the contract for watched path operations on the public API surface.
type WatchedPaths interface {
	// GetWatchedPath retrieves a watched path by ID.
	GetWatchedPath(ctx context.Context, id int64) (*models.WatchedPath, error)
	// CreateWatchedPath creates a new watched path for a tenant.
	CreateWatchedPath(ctx context.Context, tenantID int64, providerID int64, path string) (*models.WatchedPath, error)
	// UpdateWatchedPath updates an existing watched path.
	UpdateWatchedPath(ctx context.Context, id int64, path string) (*models.WatchedPath, error)
	// ListWatchedPathsByTenant retrieves watched paths for a tenant using cursor pagination.
	ListWatchedPathsByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.WatchedPath], error)
}
