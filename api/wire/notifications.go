package wire

import (
	"context"
	"encoding/json"
	"time"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

var validNotificationStatuses = []string{
	string(models.NotificationUnread),
	string(models.NotificationRead),
	string(models.NotificationDismissed),
}

// NotificationResponse is the public API response for a notification.
//
//nolint:govet // fieldalignment: readability over alignment for wire types
type NotificationResponse struct {
	CreatedAt  time.Time                `json:"created_at" description:"Creation timestamp"`
	Type       models.NotificationType  `json:"type" description:"Event type" example:"ingest.completed" discriminator:"metadata"`
	Status     models.NotificationStatus `json:"status" description:"Read status" example:"unread"`
	Metadata   json.RawMessage          `json:"metadata,omitempty" description:"Event-specific metadata" discriminate:"ProviderCreatedMeta,ProviderUpdatedMeta,ProviderConnectedMeta,ProviderDeletedMeta,DocumentIngestedMeta,WatchedPathCreatedMeta,WatchedPathUpdatedMeta,TopicCreatedMeta,TopicUpdatedMeta,TagCreatedMeta,TagUpdatedMeta,TenantCreatedMeta,TenantUpdatedMeta,TenantDeletedMeta,IngestStageMeta,IngestFailedMeta"`
	ID         string                   `json:"id" description:"Notification ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	DocumentID string                   `json:"document_id,omitempty" description:"Related document ID"`
	VersionID  string                   `json:"version_id,omitempty" description:"Related version ID"`
	Message    string                   `json:"message" description:"Notification message"`
}

// OnSend applies boundary masking before the response is marshaled.
func (n *NotificationResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[NotificationResponse]](ctx)
	masked, err := b.Send(ctx, *n)
	if err != nil {
		return err
	}
	*n = masked
	return nil
}

// Clone returns a deep copy of the response.
func (n NotificationResponse) Clone() NotificationResponse {
	c := n
	if n.Metadata != nil {
		c.Metadata = make(json.RawMessage, len(n.Metadata))
		copy(c.Metadata, n.Metadata)
	}
	return c
}

// NotificationListResponse is the public API response for a paginated notification list.
type NotificationListResponse struct {
	Notifications []NotificationResponse `json:"notifications" description:"List of notifications"`
	Offset        int                    `json:"offset" description:"Number of results skipped"`
	Limit         int                    `json:"limit" description:"Page size" example:"20"`
	Total         int64                  `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *NotificationListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[NotificationListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r NotificationListResponse) Clone() NotificationListResponse {
	c := r
	if r.Notifications != nil {
		c.Notifications = make([]NotificationResponse, len(r.Notifications))
		copy(c.Notifications, r.Notifications)
	}
	return c
}

// NotificationUpdateRequest is the request body for updating a notification's status.
type NotificationUpdateRequest struct {
	Status string `json:"status" description:"New status" example:"read"`
}

// Validate validates the request fields.
func (r *NotificationUpdateRequest) Validate() error {
	return check.All(
		check.Str(r.Status, "status").Required().OneOf(validNotificationStatuses).V(),
	).Err()
}

// Clone returns a copy of the request.
func (r NotificationUpdateRequest) Clone() NotificationUpdateRequest {
	return r
}

// NotificationBulkReadRequest is the request body for bulk-marking notifications as read.
type NotificationBulkReadRequest struct {
	Status string `json:"status" description:"Status to apply" example:"read"`
}

// Validate validates the request fields.
func (r *NotificationBulkReadRequest) Validate() error {
	return check.All(
		check.Str(r.Status, "status").Required().OneOf(validNotificationStatuses).V(),
	).Err()
}

// Clone returns a copy of the request.
func (r NotificationBulkReadRequest) Clone() NotificationBulkReadRequest {
	return r
}
