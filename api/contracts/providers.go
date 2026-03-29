package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Providers defines the contract for provider operations on the public API surface.
type Providers interface {
	// GetProvider retrieves a provider by ID.
	GetProvider(ctx context.Context, id string) (*models.Provider, error)
	// GetProviderByTenant retrieves a provider by ID scoped to a tenant.
	GetProviderByTenant(ctx context.Context, id, tenantID string) (*models.Provider, error)
	// CreateProvider creates a new provider for a tenant.
	CreateProvider(ctx context.Context, tenantID string, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
	// UpdateProvider updates an existing provider.
	UpdateProvider(ctx context.Context, id string, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
	// UpdateProviderCredentials stores credentials and marks the provider active.
	UpdateProviderCredentials(ctx context.Context, id, credentials string) error
	// ListProvidersByTenant retrieves providers for a tenant using cursor pagination.
	ListProvidersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Provider], error)
}
