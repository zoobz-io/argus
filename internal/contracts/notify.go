package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// NotifySubscriptions defines subscription lookup operations needed by the notification pipeline.
type NotifySubscriptions interface {
	// FindByTenantAndEventType retrieves subscriptions matching a tenant and event type.
	FindByTenantAndEventType(ctx context.Context, tenantID string, eventType string) ([]*models.Subscription, error)
}

// NotifyIndexer defines indexing operations needed by the notification pipeline.
type NotifyIndexer interface {
	// Index persists a notification to the search index.
	Index(ctx context.Context, notification *models.Notification) error
}

// NotifyHookLoader defines hook lookup operations needed by the webhook delivery pipeline.
type NotifyHookLoader interface {
	// GetWithSecret retrieves a hook by ID and tenant including its secret for signing.
	GetWithSecret(ctx context.Context, tenantID, id string) (*models.Hook, error)
}

// NotifyDeliveryLogger defines delivery logging operations needed by the webhook delivery pipeline.
type NotifyDeliveryLogger interface {
	// CreateDelivery records a webhook delivery attempt.
	CreateDelivery(ctx context.Context, hookID string, eventID string, tenantID string, statusCode int, attempt int, deliveryErr *string) error
}
