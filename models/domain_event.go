package models

import (
	"encoding/json"
	"time"
)

// DomainEvent is the unified source event type for audit logging, notifications,
// and webhook delivery. Every domain action (user-initiated or system) produces
// a DomainEvent that flows through a single herald stream to the notifier sidecar.
//
// The notifier routes each event to:
//   - Audit index (always — every event is an audit entry)
//   - Notification fan-out (if any subscription matches the Action)
//   - Webhook delivery (if any webhook subscription matches)
type DomainEvent struct {
	Timestamp    time.Time       `json:"timestamp"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	ID           string          `json:"id"`
	TenantID     string          `json:"tenant_id"`
	ActorID      string          `json:"actor_id"`
	Message      string          `json:"message"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
}

// Clone returns a deep copy of the domain event.
func (e DomainEvent) Clone() DomainEvent {
	c := e
	if e.Metadata != nil {
		c.Metadata = make(json.RawMessage, len(e.Metadata))
		copy(c.Metadata, e.Metadata)
	}
	return c
}

// DomainEventSearchParams holds filter parameters for domain event queries.
type DomainEventSearchParams struct {
	From         *time.Time
	To           *time.Time
	Action       string
	ResourceType string
	ActorID      string
	TenantID     string
	Offset       int
	Limit        int
}

// PageSize returns the effective page size, clamped to bounds.
func (p DomainEventSearchParams) PageSize() int {
	if p.Limit <= 0 {
		return DefaultPageSize
	}
	if p.Limit > MaxPageSize {
		return MaxPageSize
	}
	return p.Limit
}
