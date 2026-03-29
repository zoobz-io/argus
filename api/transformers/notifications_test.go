//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestNotificationToResponse(t *testing.T) {
	n := argustest.NewNotification()
	resp := NotificationToResponse(n)

	if resp.ID != "n1" || resp.Message != "Ingest completed" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.Type != models.NotificationIngestCompleted {
		t.Errorf("Type = %q, want %q", resp.Type, models.NotificationIngestCompleted)
	}
	if resp.Status != models.NotificationUnread {
		t.Errorf("Status = %q, want %q", resp.Status, models.NotificationUnread)
	}
	if resp.DocumentID != "d1" || resp.VersionID != "v1" {
		t.Errorf("document/version mismatch: %+v", resp)
	}
}

func TestNotificationsToListResponse(t *testing.T) {
	n1 := argustest.NewNotification()
	n2 := argustest.NewNotification()
	n2.ID = "n2"

	result := &models.OffsetResult[models.Notification]{
		Items:  []*models.Notification{n1, n2},
		Total:  5,
		Offset: 0,
	}

	resp := NotificationsToListResponse(result)
	if len(resp.Notifications) != 2 {
		t.Fatalf("expected 2 notifications, got %d", len(resp.Notifications))
	}
	if resp.Notifications[0].ID != "n1" || resp.Notifications[1].ID != "n2" {
		t.Errorf("IDs mismatch: %q, %q", resp.Notifications[0].ID, resp.Notifications[1].ID)
	}
	if resp.Total != 5 {
		t.Errorf("Total = %d, want 5", resp.Total)
	}
	if resp.Limit != 2 {
		t.Errorf("Limit = %d, want 2", resp.Limit)
	}
}

func TestNotificationsToListResponse_Empty(t *testing.T) {
	result := &models.OffsetResult[models.Notification]{
		Items: []*models.Notification{},
		Total: 0,
	}
	resp := NotificationsToListResponse(result)
	if len(resp.Notifications) != 0 {
		t.Errorf("expected empty, got %d", len(resp.Notifications))
	}
}
