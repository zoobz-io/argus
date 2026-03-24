package stores

import (
	"context"
	"fmt"

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

// GetWatchedPath retrieves a watched path by ID.
func (s *WatchedPaths) GetWatchedPath(ctx context.Context, id int64) (*models.WatchedPath, error) {
	return s.Select().
		Where("id", "=", ":id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateWatchedPath creates a new watched path for a tenant.
func (s *WatchedPaths) CreateWatchedPath(ctx context.Context, tenantID int64, providerID int64, path string) (*models.WatchedPath, error) {
	wp := &models.WatchedPath{
		TenantID:   tenantID,
		ProviderID: providerID,
		Path:       path,
		Active:     true,
	}
	if err := s.Set(ctx, "", wp); err != nil {
		return nil, fmt.Errorf("creating watched path: %w", err)
	}
	return wp, nil
}

// UpdateWatchedPath updates an existing watched path.
func (s *WatchedPaths) UpdateWatchedPath(ctx context.Context, id int64, path string) (*models.WatchedPath, error) {
	wp, err := s.GetWatchedPath(ctx, id)
	if err != nil {
		return nil, err
	}
	wp.Path = path
	if err := s.Set(ctx, fmt.Sprintf("%d", id), wp); err != nil {
		return nil, fmt.Errorf("updating watched path: %w", err)
	}
	return wp, nil
}

// DeleteWatchedPath removes a watched path.
func (s *WatchedPaths) DeleteWatchedPath(ctx context.Context, id int64) error {
	return s.Delete(ctx, fmt.Sprintf("%d", id))
}

// ListWatchedPaths retrieves a paginated list of all watched paths.
func (s *WatchedPaths) ListWatchedPaths(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.WatchedPath], error) {
	limit := page.PageSize()
	q := s.Query().OrderBy("id", "ASC").Limit(limit + 1)
	params := map[string]any{}
	if page.Cursor != nil {
		q = q.Where("id", ">", ":cursor")
		params["cursor"] = *page.Cursor
	}
	items, err := q.Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	return cursorResult(items, limit), nil
}

// ListWatchedPathsByTenant retrieves watched paths for a specific tenant using cursor pagination.
func (s *WatchedPaths) ListWatchedPathsByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.WatchedPath], error) {
	limit := page.PageSize()
	q := s.Query().Where("tenant_id", "=", ":tenant_id").OrderBy("id", "ASC").Limit(limit + 1)
	params := map[string]any{"tenant_id": tenantID}
	if page.Cursor != nil {
		q = q.Where("id", ">", ":cursor")
		params["cursor"] = *page.Cursor
	}
	items, err := q.Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	return cursorResult(items, limit), nil
}
