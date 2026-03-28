package models

import "time"

// NotificationType identifies the kind of notification.
type NotificationType string

// Notification type constants.
const (
	NotificationIngestCompleted NotificationType = "ingest.completed"
	NotificationIngestFailed    NotificationType = "ingest.failed"
)

// NotificationStatus tracks whether a notification has been seen.
type NotificationStatus string

// Notification status constants.
const (
	NotificationUnread    NotificationStatus = "unread"
	NotificationRead      NotificationStatus = "read"
	NotificationDismissed NotificationStatus = "dismissed"
)

// Notification is the OpenSearch document type for application notifications.
type Notification struct {
	CreatedAt  time.Time          `json:"created_at"`
	Type       NotificationType   `json:"type"`
	Status     NotificationStatus `json:"status"`
	ID         string             `json:"id"`
	TenantID   string             `json:"tenant_id"`
	DocumentID string             `json:"document_id,omitempty"`
	VersionID  string             `json:"version_id,omitempty"`
	Message    string             `json:"message"`
	Error      string             `json:"error,omitempty"`
}

// Clone returns a deep copy of the notification.
func (n Notification) Clone() Notification {
	return n
}
