package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Hooks defines admin operations for webhook endpoints.
type Hooks interface {
	// GetHook retrieves a webhook endpoint by ID.
	GetHook(ctx context.Context, id string) (*models.Hook, error)
	// ListHooks retrieves a paginated list of all webhook endpoints (cross-tenant).
	ListHooks(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Hook], error)
	// DeleteHook removes a webhook endpoint.
	DeleteHook(ctx context.Context, id string) error
	// ListDeliveries retrieves a paginated list of all deliveries (cross-tenant).
	ListDeliveries(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Delivery], error)
}
