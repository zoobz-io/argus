//go:build testing

package wire

import (
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
)

func TestNotificationResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := NotificationResponse{
		ID: "n1", Type: models.NotificationIngestCompleted,
		Status: models.NotificationUnread, Message: "Done",
		CreatedAt: time.Now(),
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "n1" {
		t.Errorf("ID mutated: %q", resp.ID)
	}
}

func TestNotificationResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := NotificationResponse{ID: "n1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationResponse_Clone(t *testing.T) {
	orig := NotificationResponse{ID: "n1", Message: "Done", Status: models.NotificationUnread}
	cloned := orig.Clone()
	if cloned.ID != orig.ID || cloned.Message != orig.Message {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestNotificationListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := NotificationListResponse{
		Notifications: []NotificationResponse{{ID: "n1"}},
		Limit:         20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotificationListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := NotificationListResponse{Notifications: []NotificationResponse{{ID: "n1"}}}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestNotificationListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := NotificationListResponse{
		Notifications: []NotificationResponse{{ID: "n1", Message: "Done"}},
		Limit:         20,
	}
	cloned := orig.Clone()

	cloned.Notifications[0].Message = "mutated"
	if orig.Notifications[0].Message == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestNotificationListResponse_Clone_NilNotifications(t *testing.T) {
	orig := NotificationListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Notifications != nil {
		t.Error("expected nil Notifications")
	}
}

func TestNotificationUpdateRequest_Validate_Valid(t *testing.T) {
	req := NotificationUpdateRequest{Status: "read"}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestNotificationUpdateRequest_Validate_MissingStatus(t *testing.T) {
	req := NotificationUpdateRequest{}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty status")
	}
}

func TestNotificationUpdateRequest_Validate_InvalidStatus(t *testing.T) {
	req := NotificationUpdateRequest{Status: "invalid"}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for invalid status")
	}
}

func TestNotificationUpdateRequest_Clone(t *testing.T) {
	orig := NotificationUpdateRequest{Status: "read"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestNotificationBulkReadRequest_Validate_Valid(t *testing.T) {
	req := NotificationBulkReadRequest{Status: "read"}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestNotificationBulkReadRequest_Validate_MissingStatus(t *testing.T) {
	req := NotificationBulkReadRequest{}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty status")
	}
}

func TestNotificationBulkReadRequest_Validate_InvalidStatus(t *testing.T) {
	req := NotificationBulkReadRequest{Status: "bogus"}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for invalid status")
	}
}

func TestNotificationBulkReadRequest_Clone(t *testing.T) {
	orig := NotificationBulkReadRequest{Status: "read"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}
