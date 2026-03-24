package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Providers defines admin operations for providers.
type Providers interface {
	// GetProvider retrieves a provider by ID.
	GetProvider(ctx context.Context, id int64) (*models.Provider, error)
	// CreateProvider creates a new provider for a tenant.
	CreateProvider(ctx context.Context, tenantID int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
	// UpdateProvider updates an existing provider.
	UpdateProvider(ctx context.Context, id int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
	// DeleteProvider removes a provider.
	DeleteProvider(ctx context.Context, id int64) error
	// ListProviders retrieves a paginated list of all providers.
	ListProviders(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Provider], error)
	// ListProvidersByTenant retrieves providers for a specific tenant.
	ListProvidersByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Provider], error)
}
