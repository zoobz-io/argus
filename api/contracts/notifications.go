package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Notifications defines the contract for notification operations on the public API surface.
type Notifications interface {
	// SearchByUser searches notifications for a user within a tenant.
	SearchByUser(ctx context.Context, tenantID string, userID string, page models.OffsetPage) (*models.OffsetResult[models.Notification], error)
	// UpdateStatus updates the status of a single notification.
	UpdateStatus(ctx context.Context, tenantID string, id string, status models.NotificationStatus) (*models.Notification, error)
	// BulkUpdateStatus marks all notifications for a user as the given status.
	BulkUpdateStatus(ctx context.Context, tenantID string, userID string, status models.NotificationStatus) error
}
