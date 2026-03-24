package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Tenants defines the contract for tenant operations on the public API surface.
type Tenants interface {
	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, id int64) (*models.Tenant, error)
	// ListTenants retrieves a paginated list of tenants.
	ListTenants(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Tenant], error)
}
