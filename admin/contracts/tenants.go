package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Tenants defines admin operations for tenants.
type Tenants interface {
	// GetTenant retrieves a tenant by ID.
	GetTenant(ctx context.Context, id string) (*models.Tenant, error)
	// CreateTenant creates a new tenant.
	CreateTenant(ctx context.Context, name string, slug string) (*models.Tenant, error)
	// UpdateTenant updates an existing tenant.
	UpdateTenant(ctx context.Context, id string, name string, slug string) (*models.Tenant, error)
	// DeleteTenant removes a tenant.
	DeleteTenant(ctx context.Context, id string) error
	// ListTenants retrieves a paginated list of all tenants.
	ListTenants(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Tenant], error)
}
