package models

import (
	"encoding/json"
	"time"
)

// SubscriptionChannel identifies the delivery channel for a subscription.
type SubscriptionChannel string

// Subscription channel constants.
const (
	SubscriptionChannelInbox   SubscriptionChannel = "inbox"
	SubscriptionChannelWebhook SubscriptionChannel = "webhook"
)

// Subscription represents a user's opt-in to receive notifications for a specific event type.
type Subscription struct {
	CreatedAt         time.Time           `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt         time.Time           `json:"updated_at" db:"updated_at" default:"now()"`
	WebhookEndpointID *string             `json:"webhook_endpoint_id,omitempty" db:"webhook_endpoint_id"`
	Channel           SubscriptionChannel `json:"channel" db:"channel" constraints:"notnull" default:"'inbox'"`
	ID                string              `json:"id" db:"id" constraints:"primarykey"`
	UserID            string              `json:"user_id" db:"user_id" constraints:"notnull"`
	TenantID          string              `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	EventType         string              `json:"event_type" db:"event_type" constraints:"notnull"`
	Filters           json.RawMessage     `json:"filters,omitempty" db:"filters"`
}

// GetID returns the subscription's primary key.
func (s Subscription) GetID() string {
	return s.ID
}

// GetCreatedAt returns the subscription's creation timestamp.
func (s Subscription) GetCreatedAt() time.Time {
	return s.CreatedAt
}

// Clone returns a deep copy of the subscription.
func (s Subscription) Clone() Subscription {
	c := s
	if s.Filters != nil {
		f := make(json.RawMessage, len(s.Filters))
		copy(f, s.Filters)
		c.Filters = f
	}
	if s.WebhookEndpointID != nil {
		w := *s.WebhookEndpointID
		c.WebhookEndpointID = &w
	}
	return c
}
