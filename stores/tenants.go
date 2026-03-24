package stores

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// Tenants provides database access for tenants.
type Tenants struct {
	*sum.Database[models.Tenant]
}

// NewTenants creates a new tenants store.
func NewTenants(db *sqlx.DB, renderer astql.Renderer) (*Tenants, error) {
	database, err := sum.NewDatabase[models.Tenant](db, "tenants", renderer)
	if err != nil {
		return nil, err
	}
	return &Tenants{Database: database}, nil
}

// GetTenant retrieves a tenant by ID.
func (s *Tenants) GetTenant(ctx context.Context, id int64) (*models.Tenant, error) {
	return s.Select().
		Where("id", "=", ":id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateTenant creates a new tenant.
func (s *Tenants) CreateTenant(ctx context.Context, name string, slug string) (*models.Tenant, error) {
	t := &models.Tenant{
		Name: name,
		Slug: slug,
	}
	if err := s.Set(ctx, "", t); err != nil {
		return nil, fmt.Errorf("creating tenant: %w", err)
	}
	return t, nil
}

// UpdateTenant updates an existing tenant.
func (s *Tenants) UpdateTenant(ctx context.Context, id int64, name string, slug string) (*models.Tenant, error) {
	t, err := s.GetTenant(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	t.Slug = slug
	if err := s.Set(ctx, fmt.Sprintf("%d", id), t); err != nil {
		return nil, fmt.Errorf("updating tenant: %w", err)
	}
	return t, nil
}

// DeleteTenant removes a tenant.
func (s *Tenants) DeleteTenant(ctx context.Context, id int64) error {
	return s.Delete(ctx, fmt.Sprintf("%d", id))
}

// ListTenants retrieves a paginated list of tenants using cursor-based pagination.
func (s *Tenants) ListTenants(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Tenant], error) {
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
