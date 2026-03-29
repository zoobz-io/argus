//go:build testing

package notify

import (
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestFanOutItem_Clone_Full(t *testing.T) {
	item := &FanOutItem{
		EventID: "evt-1",
		Notification: &models.Notification{
			ID:       "n-1",
			UserID:   "u-1",
			TenantID: "t-1",
			Type:     models.NotificationIngestCompleted,
			Message:  "done",
		},
		Subscription: &models.Subscription{
			ID:     "sub-1",
			UserID: "u-1",
		},
	}

	cloned := item.Clone()

	if cloned.EventID != "evt-1" {
		t.Errorf("EventID: got %q, want %q", cloned.EventID, "evt-1")
	}
	if cloned.Notification == nil {
		t.Fatal("expected cloned Notification to be non-nil")
	}
	if cloned.Notification.ID != "n-1" {
		t.Errorf("Notification.ID: got %q, want %q", cloned.Notification.ID, "n-1")
	}
	if cloned.Subscription == nil {
		t.Fatal("expected cloned Subscription to be non-nil")
	}
	if cloned.Subscription.ID != "sub-1" {
		t.Errorf("Subscription.ID: got %q, want %q", cloned.Subscription.ID, "sub-1")
	}

	// Verify deep copy — mutating original should not affect clone.
	item.Notification.ID = "mutated"
	item.Subscription.ID = "mutated"
	if cloned.Notification.ID == "mutated" {
		t.Error("clone Notification should not be affected by mutation of original")
	}
	if cloned.Subscription.ID == "mutated" {
		t.Error("clone Subscription should not be affected by mutation of original")
	}
}

func TestFanOutItem_Clone_NilNotification(t *testing.T) {
	item := &FanOutItem{
		EventID: "evt-1",
		Subscription: &models.Subscription{
			ID:     "sub-1",
			UserID: "u-1",
		},
	}

	cloned := item.Clone()

	if cloned.Notification != nil {
		t.Error("expected cloned Notification to be nil")
	}
	if cloned.Subscription == nil {
		t.Fatal("expected cloned Subscription to be non-nil")
	}
	if cloned.EventID != "evt-1" {
		t.Errorf("EventID: got %q, want %q", cloned.EventID, "evt-1")
	}
}

func TestFanOutItem_Clone_NilSubscription(t *testing.T) {
	item := &FanOutItem{
		EventID: "evt-1",
		Notification: &models.Notification{
			ID:     "n-1",
			UserID: "u-1",
		},
	}

	cloned := item.Clone()

	if cloned.Subscription != nil {
		t.Error("expected cloned Subscription to be nil")
	}
	if cloned.Notification == nil {
		t.Fatal("expected cloned Notification to be non-nil")
	}
}

func TestFanOutItem_Clone_BothNil(t *testing.T) {
	item := &FanOutItem{
		EventID: "evt-1",
	}

	cloned := item.Clone()

	if cloned.EventID != "evt-1" {
		t.Errorf("EventID: got %q, want %q", cloned.EventID, "evt-1")
	}
	if cloned.Notification != nil {
		t.Error("expected Notification to be nil")
	}
	if cloned.Subscription != nil {
		t.Error("expected Subscription to be nil")
	}
}
