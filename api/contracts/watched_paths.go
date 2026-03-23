package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// WatchedPaths defines the contract for watched path data access.
type WatchedPaths interface {
	// Get retrieves a watched path by primary key.
	Get(ctx context.Context, key string) (*models.WatchedPath, error)
	// Set creates or updates a watched path.
	Set(ctx context.Context, key string, path *models.WatchedPath) error
}
