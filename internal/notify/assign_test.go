//go:build testing

package notify

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
)

func TestAssignStage_SetsAllFields(t *testing.T) {
	stage := newAssignStage()

	item := &FanOutItem{
		Notification: &models.Notification{
			TenantID: "t-1",
			Type:     models.NotificationIngestCompleted,
			Message:  "Document ingestion completed",
		},
		Subscription: &models.Subscription{
			ID:     "sub-1",
			UserID: "u-1",
		},
		EventID: "evt-1",
	}

	before := time.Now()
	result, err := stage.Process(context.Background(), item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Notification.ID == "" {
		t.Error("expected notification ID to be set")
	}
	if result.Notification.UserID != "u-1" {
		t.Errorf("UserID: got %q, want %q", result.Notification.UserID, "u-1")
	}
	if result.Notification.EventID != "evt-1" {
		t.Errorf("EventID: got %q, want %q", result.Notification.EventID, "evt-1")
	}
	if result.Notification.Status != models.NotificationUnread {
		t.Errorf("Status: got %q, want %q", result.Notification.Status, models.NotificationUnread)
	}
	if result.Notification.CreatedAt.Before(before) {
		t.Error("expected CreatedAt to be set to current time")
	}
}
