package stores

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// Providers provides database access for providers.
type Providers struct {
	*sum.Database[models.Provider]
}

// NewProviders creates a new providers store.
func NewProviders(db *sqlx.DB, renderer astql.Renderer) (*Providers, error) {
	database, err := sum.NewDatabase[models.Provider](db, "providers", renderer)
	if err != nil {
		return nil, err
	}
	return &Providers{Database: database}, nil
}

// GetProvider retrieves a provider by ID.
func (s *Providers) GetProvider(ctx context.Context, id int64) (*models.Provider, error) {
	return s.Select().
		Where("id", "=", ":id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateProvider creates a new provider for a tenant.
func (s *Providers) CreateProvider(ctx context.Context, tenantID int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error) {
	p := &models.Provider{
		TenantID:    tenantID,
		Type:        providerType,
		Name:        name,
		Credentials: credentials,
		Active:      true,
	}
	if err := s.Set(ctx, "", p); err != nil {
		return nil, fmt.Errorf("creating provider: %w", err)
	}
	return p, nil
}

// UpdateProvider updates an existing provider.
func (s *Providers) UpdateProvider(ctx context.Context, id int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error) {
	p, err := s.GetProvider(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Type = providerType
	p.Name = name
	p.Credentials = credentials
	if err := s.Set(ctx, fmt.Sprintf("%d", id), p); err != nil {
		return nil, fmt.Errorf("updating provider: %w", err)
	}
	return p, nil
}

// DeleteProvider removes a provider.
func (s *Providers) DeleteProvider(ctx context.Context, id int64) error {
	return s.Delete(ctx, fmt.Sprintf("%d", id))
}

// ListProviders retrieves a paginated list of all providers.
func (s *Providers) ListProviders(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Provider], error) {
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

// ListProvidersByTenant retrieves providers for a specific tenant using cursor pagination.
func (s *Providers) ListProvidersByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Provider], error) {
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
