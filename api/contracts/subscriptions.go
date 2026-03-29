package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Subscriptions defines the contract for subscription operations on the public API surface.
type Subscriptions interface {
	// GetSubscriptionByTenant retrieves a subscription by ID scoped to a tenant.
	GetSubscriptionByTenant(ctx context.Context, tenantID string, id string) (*models.Subscription, error)
	// ListSubscriptionsByUser retrieves subscriptions for a user within a tenant.
	ListSubscriptionsByUser(ctx context.Context, tenantID string, userID string, page models.OffsetPage) (*models.OffsetResult[models.Subscription], error)
	// CreateSubscription creates a new subscription for a user.
	CreateSubscription(ctx context.Context, tenantID string, userID string, eventType string, channel models.SubscriptionChannel) (*models.Subscription, error)
	// DeleteSubscription removes a subscription by ID scoped to tenant and user.
	DeleteSubscription(ctx context.Context, tenantID string, userID string, id string) error
}
