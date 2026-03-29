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
