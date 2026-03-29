//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestSubscriptionToResponse(t *testing.T) {
	sub := argustest.NewSubscription()
	resp := SubscriptionToResponse(sub)

	if resp.ID != "s1" || resp.UserID != "u1" || resp.EventType != "ingest.completed" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.Channel != models.SubscriptionChannelInbox {
		t.Errorf("Channel = %q, want %q", resp.Channel, models.SubscriptionChannelInbox)
	}
}

func TestSubscriptionsToResponse(t *testing.T) {
	subs := []*models.Subscription{argustest.NewSubscription(), argustest.NewSubscription()}
	subs[1].ID = "s2"

	result := SubscriptionsToResponse(subs)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].ID != "s1" || result[1].ID != "s2" {
		t.Errorf("IDs mismatch: %q, %q", result[0].ID, result[1].ID)
	}
}

func TestSubscriptionsToResponse_Empty(t *testing.T) {
	result := SubscriptionsToResponse(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}
