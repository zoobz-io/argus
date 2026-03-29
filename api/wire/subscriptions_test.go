//go:build testing

package wire

import (
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestSubscriptionRequest_Validate_Valid(t *testing.T) {
	req := SubscriptionRequest{EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestSubscriptionRequest_Validate_MissingEventType(t *testing.T) {
	req := SubscriptionRequest{Channel: models.SubscriptionChannelInbox}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty event_type")
	}
}

func TestSubscriptionRequest_Validate_MissingChannel(t *testing.T) {
	req := SubscriptionRequest{EventType: "ingest.completed"}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty channel")
	}
}

func TestSubscriptionRequest_Clone(t *testing.T) {
	orig := SubscriptionRequest{EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestSubscriptionResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := SubscriptionResponse{ID: "s1", UserID: "u1", EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "s1" || resp.EventType != "ingest.completed" {
		t.Errorf("fields mutated: %+v", resp)
	}
}

func TestSubscriptionResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := SubscriptionResponse{ID: "s1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestSubscriptionResponse_Clone(t *testing.T) {
	orig := SubscriptionResponse{ID: "s1", UserID: "u1", EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestSubscriptionListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := SubscriptionListResponse{
		Subscriptions: []SubscriptionResponse{{ID: "s1"}},
		Limit:         20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSubscriptionListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := SubscriptionListResponse{Subscriptions: []SubscriptionResponse{{ID: "s1"}}}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestSubscriptionListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := SubscriptionListResponse{
		Subscriptions: []SubscriptionResponse{{ID: "s1", EventType: "ingest.completed"}},
		Limit:         20,
	}
	cloned := orig.Clone()

	cloned.Subscriptions[0].EventType = "mutated"
	if orig.Subscriptions[0].EventType == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestSubscriptionListResponse_Clone_NilSubscriptions(t *testing.T) {
	orig := SubscriptionListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Subscriptions != nil {
		t.Error("expected nil Subscriptions")
	}
}
