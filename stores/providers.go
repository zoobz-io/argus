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

// Providers provides database access for providers.
type Providers struct {
	*sum.Database[models.Provider]
}

// NewProviders creates a new providers store.
func NewProviders(db *sqlx.DB, renderer astql.Renderer) *Providers {
	return &Providers{
		Database: sum.NewDatabase[models.Provider](db, "providers", renderer),
	}
}

// GetProvider retrieves a provider by ID.
func (s *Providers) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateProvider creates a new provider for a tenant.
func (s *Providers) CreateProvider(ctx context.Context, tenantID string, providerType models.ProviderType, name string, credentials string) (*models.Provider, error) {
	p := &models.Provider{
		ID:          uuid.New().String(),
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
func (s *Providers) UpdateProvider(ctx context.Context, id string, providerType models.ProviderType, name string, credentials string) (*models.Provider, error) {
	p, err := s.GetProvider(ctx, id)
	if err != nil {
		return nil, err
	}
	p.Type = providerType
	p.Name = name
	p.Credentials = credentials
	if err := s.Set(ctx, id, p); err != nil {
		return nil, fmt.Errorf("updating provider: %w", err)
	}
	return p, nil
}

// DeleteProvider removes a provider.
func (s *Providers) DeleteProvider(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ListProviders retrieves a paginated list of all providers.
func (s *Providers) ListProviders(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().Exec(ctx, nil)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Provider]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}

// ListProvidersByTenant retrieves providers for a specific tenant using offset/limit pagination.
func (s *Providers) ListProvidersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
	params := map[string]any{"tenant_id": tenantID}
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, params)
	if err != nil {
		return nil, err
	}
	total, countErr := s.Count().
		Where("tenant_id", "=", "tenant_id").
		Exec(ctx, params)
	if countErr != nil {
		return nil, countErr
	}
	return &models.OffsetResult[models.Provider]{Items: items, Total: int64(total), Offset: page.Offset}, nil
}
