package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Tenants defines the contract for tenant data access.
type Tenants interface {
	// Get retrieves a tenant by primary key.
	Get(ctx context.Context, key string) (*models.Tenant, error)
	// Set creates or updates a tenant.
	Set(ctx context.Context, key string, tenant *models.Tenant) error
}
