package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Providers defines admin operations for providers.
type Providers interface {
	// GetProvider retrieves a provider by ID.
	GetProvider(ctx context.Context, id string) (*models.Provider, error)
	// CreateProvider creates a new provider for a tenant.
	CreateProvider(ctx context.Context, tenantID string, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
	// UpdateProvider updates an existing provider scoped to a tenant.
	UpdateProvider(ctx context.Context, tenantID, id string, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
	// DeleteProvider removes a provider.
	DeleteProvider(ctx context.Context, id string) error
	// ListProviders retrieves a paginated list of all providers.
	ListProviders(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
	// ListProvidersByTenant retrieves providers for a specific tenant.
	ListProvidersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
}
