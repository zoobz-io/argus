package models

import (
	"encoding/json"
	"time"
)

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
// Materialized from DomainEvent during subscription fan-out — carries per-user
// read/unread state and typed event metadata.
//
//nolint:govet // fieldalignment: readability over alignment for domain types
type Notification struct {
	CreatedAt  time.Time          `json:"created_at"`
	Type       NotificationType   `json:"type"`
	Status     NotificationStatus `json:"status"`
	Metadata   json.RawMessage    `json:"metadata,omitempty"`
	ID         string             `json:"id"`
	UserID     string             `json:"user_id"`
	EventID    string             `json:"event_id,omitempty"`
	TenantID   string             `json:"tenant_id"`
	DocumentID string             `json:"document_id,omitempty"`
	VersionID  string             `json:"version_id,omitempty"`
	Message    string             `json:"message"`
	Error      string             `json:"error,omitempty"`
}

// Clone returns a deep copy of the notification.
func (n Notification) Clone() Notification {
	c := n
	if n.Metadata != nil {
		c.Metadata = make(json.RawMessage, len(n.Metadata))
		copy(c.Metadata, n.Metadata)
	}
	return c
}

// NotificationSearchResult holds search results for notifications.
type NotificationSearchResult struct {
	Items  []*Notification
	Total  int64
	Offset int
}
