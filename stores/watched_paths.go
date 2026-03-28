package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
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
func NewWatchedPaths(db *sqlx.DB, renderer astql.Renderer) *WatchedPaths {
	return &WatchedPaths{
		Database: sum.NewDatabase[models.WatchedPath](db, "watched_paths", renderer),
	}
}

// GetWatchedPath retrieves a watched path by ID.
func (s *WatchedPaths) GetWatchedPath(ctx context.Context, id string) (*models.WatchedPath, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateWatchedPath creates a new watched path for a tenant.
func (s *WatchedPaths) CreateWatchedPath(ctx context.Context, tenantID string, providerID string, path string) (*models.WatchedPath, error) {
	wp := &models.WatchedPath{
		ID:         uuid.New().String(),
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
func (s *WatchedPaths) UpdateWatchedPath(ctx context.Context, id string, path string) (*models.WatchedPath, error) {
	wp, err := s.GetWatchedPath(ctx, id)
	if err != nil {
		return nil, err
	}
	wp.Path = path
	if err := s.Set(ctx, id, wp); err != nil {
		return nil, fmt.Errorf("updating watched path: %w", err)
	}
	return wp, nil
}

// DeleteWatchedPath removes a watched path.
func (s *WatchedPaths) DeleteWatchedPath(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ListWatchedPaths retrieves a paginated list of all watched paths.
func (s *WatchedPaths) ListWatchedPaths(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.WatchedPath]{Items: items, Offset: page.Offset}, nil
}

// ListWatchedPathsByTenant retrieves watched paths for a specific tenant using offset/limit pagination.
func (s *WatchedPaths) ListWatchedPathsByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.WatchedPath]{Items: items, Offset: page.Offset}, nil
}
