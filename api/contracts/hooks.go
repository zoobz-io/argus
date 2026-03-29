package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Hooks defines the contract for webhook endpoint operations on the public API surface.
type Hooks interface {
	// CreateHook creates a new webhook endpoint for a tenant.
	CreateHook(ctx context.Context, tenantID string, userID string, url string) (*models.Hook, error)
	// GetHookByTenant retrieves a webhook endpoint by ID scoped to a tenant.
	GetHookByTenant(ctx context.Context, tenantID string, id string) (*models.Hook, error)
	// ListHooksByTenant retrieves a paginated list of webhook endpoints for a tenant.
	ListHooksByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Hook], error)
	// DeleteHook removes a webhook endpoint by ID scoped to a tenant.
	DeleteHook(ctx context.Context, tenantID string, id string) error
}

// Deliveries defines the contract for webhook delivery log queries.
type Deliveries interface {
	// ListByHook retrieves a paginated list of deliveries for a hook scoped to a tenant.
	ListByHook(ctx context.Context, tenantID string, hookID string, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error)
}
