package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// WatchedPaths provides database access for watched paths.
type WatchedPaths struct {
	*sum.Database[models.WatchedPath]
}

// NewWatchedPaths creates a new watched paths store.
func NewWatchedPaths(db *sqlx.DB, renderer astql.Renderer) (*WatchedPaths, error) {
	database, err := sum.NewDatabase[models.WatchedPath](db, "watched_paths", renderer)
	if err != nil {
		return nil, err
	}
	return &WatchedPaths{Database: database}, nil
}
