package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Subscriptions defines admin operations for subscriptions.
type Subscriptions interface {
	// GetSubscription retrieves a subscription by ID.
	GetSubscription(ctx context.Context, id string) (*models.Subscription, error)
	// ListSubscriptions retrieves a paginated list of all subscriptions (cross-tenant).
	ListSubscriptions(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error)
	// DeleteSubscription removes a subscription.
	DeleteSubscription(ctx context.Context, id string) error
}
