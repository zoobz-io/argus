//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// =============================================================================
// Notifications
// =============================================================================

func TestAPI_Notifications_List(t *testing.T) {
	// Initially the test user should have no notifications (or only ones from prior tests).
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/notifications", nil)
	rtesting.AssertStatus(t, list, 200)

	var resp wire.NotificationListResponse
	if err := list.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The list endpoint should return successfully even if empty.
	t.Logf("notifications list returned %d items", len(resp.Notifications))
}

func TestAPI_Notifications_MarkRead(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Index a test notification directly so we have something to mark read.
	n := &models.Notification{
		ID:        uuid.New().String(),
		UserID:    testUserID,
		TenantID:  testTenantID,
		Type:      models.NotificationIngestCompleted,
		Status:    models.NotificationUnread,
		Message:   "Test notification for mark-read",
		CreatedAt: time.Now(),
	}
	if err := s.Notifications.Index(ctx, n); err != nil {
		t.Fatalf("indexing test notification: %v", err)
	}
	RefreshNotificationsIndex(ctx, t, testInfra.OpenSearchAddr)

	// Verify it shows up in the list.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/notifications", nil)
	rtesting.AssertStatus(t, list, 200)

	var listResp wire.NotificationListResponse
	if err := list.DecodeJSON(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Notifications) == 0 {
		t.Fatal("expected at least one notification after indexing")
	}

	// Bulk mark all as read.
	body := wire.NotificationBulkReadRequest{Status: "read"}
	mark := rtesting.ServeRequest(testAPIEngine, "POST", "/notifications/mark-read", body)
	rtesting.AssertStatus(t, mark, 200)
}

// =============================================================================
// Subscriptions
// =============================================================================

func TestAPI_Subscriptions_CRUD(t *testing.T) {
	// Create.
	body := wire.SubscriptionRequest{
		EventType: "ingest.completed",
		Channel:   models.SubscriptionChannelInbox,
	}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/subscriptions", body)
	rtesting.AssertStatus(t, create, 200)

	var created wire.SubscriptionResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty subscription ID")
	}
	if created.EventType != "ingest.completed" {
		t.Errorf("event type mismatch: got %q", created.EventType)
	}
	if created.Channel != models.SubscriptionChannelInbox {
		t.Errorf("channel mismatch: got %q", created.Channel)
	}

	// List.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/subscriptions", nil)
	rtesting.AssertStatus(t, list, 200)

	var listResp wire.SubscriptionListResponse
	if err := list.DecodeJSON(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Subscriptions) == 0 {
		t.Error("expected at least one subscription")
	}

	found := false
	for _, s := range listResp.Subscriptions {
		if s.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created subscription %s not found in list", created.ID)
	}

	// Delete.
	del := rtesting.ServeRequest(testAPIEngine, "DELETE", "/subscriptions/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify it's gone from the list.
	listAfter := rtesting.ServeRequest(testAPIEngine, "GET", "/subscriptions", nil)
	rtesting.AssertStatus(t, listAfter, 200)

	var listAfterResp wire.SubscriptionListResponse
	if err := listAfter.DecodeJSON(&listAfterResp); err != nil {
		t.Fatalf("decode list after delete: %v", err)
	}
	for _, s := range listAfterResp.Subscriptions {
		if s.ID == created.ID {
			t.Errorf("subscription %s should have been deleted", created.ID)
		}
	}
}

func TestAPI_Subscriptions_DeleteNotFound(t *testing.T) {
	del := rtesting.ServeRequest(testAPIEngine, "DELETE", "/subscriptions/nonexistent", nil)
	rtesting.AssertStatus(t, del, 404)
}
