package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSubscription_GetID(t *testing.T) {
	s := Subscription{ID: "sub-1"}
	if got := s.GetID(); got != "sub-1" {
		t.Errorf("GetID() = %q, want %q", got, "sub-1")
	}
}

func TestSubscription_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	s := Subscription{CreatedAt: ts}
	if got := s.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestSubscription_Clone(t *testing.T) {
	s := Subscription{
		ID:        "sub-1",
		UserID:    "u-1",
		TenantID:  "t-1",
		EventType: "ingest.completed",
		Channel:   SubscriptionChannelInbox,
		Filters:   json.RawMessage(`{"doc_type":"pdf"}`),
	}
	clone := s.Clone()

	if clone.ID != s.ID || clone.UserID != s.UserID || clone.EventType != s.EventType {
		t.Error("Clone did not copy all fields")
	}

	// Mutating the Filters slice in clone should not affect original.
	clone.Filters[0] = '!'
	if s.Filters[0] == '!' {
		t.Error("mutating clone Filters leaked to original")
	}
}

func TestSubscription_Clone_NilFilters(t *testing.T) {
	s := Subscription{ID: "sub-2"}
	clone := s.Clone()
	if clone.Filters != nil {
		t.Error("expected nil Filters on clone")
	}
}

func TestSubscription_Clone_WebhookEndpointID(t *testing.T) {
	hookID := "h-1"
	s := Subscription{
		ID:                "sub-3",
		Channel:           SubscriptionChannelWebhook,
		WebhookEndpointID: &hookID,
	}
	clone := s.Clone()

	if clone.WebhookEndpointID == nil || *clone.WebhookEndpointID != "h-1" {
		t.Error("Clone did not copy WebhookEndpointID")
	}

	*clone.WebhookEndpointID = "mutated"
	if *s.WebhookEndpointID != "h-1" {
		t.Error("mutating clone WebhookEndpointID leaked to original")
	}
}

func TestSubscription_Clone_NilWebhookEndpointID(t *testing.T) {
	s := Subscription{ID: "sub-4"}
	clone := s.Clone()
	if clone.WebhookEndpointID != nil {
		t.Error("expected nil WebhookEndpointID on clone")
	}
}
