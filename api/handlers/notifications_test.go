//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestListNotifications_Success(t *testing.T) {
	users := defaultUsersMock()
	notifs := &argustest.MockNotifications{
		OnSearchByUser: func(_ context.Context, tenantID, userID string, _ models.OffsetPage) (*models.OffsetResult[models.Notification], error) {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %q", tenantID)
			}
			if userID != "u-1" {
				t.Errorf("expected u-1, got %q", userID)
			}
			return &models.OffsetResult[models.Notification]{
				Items: []*models.Notification{{
					ID:      "n-1",
					Type:    models.NotificationIngestCompleted,
					Status:  models.NotificationUnread,
					Message: "Done",
				}},
				Total: 1,
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithAPINotifications(notifs),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/notifications", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.NotificationListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Notifications) != 1 || resp.Notifications[0].ID != "n-1" {
		t.Errorf("unexpected notifications: %+v", resp.Notifications)
	}
}

func TestListNotifications_UserNotFound(t *testing.T) {
	users := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithAPINotifications(&argustest.MockNotifications{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/notifications", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListNotifications_Error(t *testing.T) {
	users := defaultUsersMock()
	notifs := &argustest.MockNotifications{
		OnSearchByUser: func(_ context.Context, _, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Notification], error) {
			return nil, fmt.Errorf("search error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithAPINotifications(notifs),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	capture := rtesting.ServeRequest(engine, "GET", "/notifications", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestUpdateNotificationStatus_Success(t *testing.T) {
	notifs := &argustest.MockNotifications{
		OnUpdateStatus: func(_ context.Context, _, _, id string, status models.NotificationStatus) (*models.Notification, error) {
			return &models.Notification{
				ID:        id,
				Type:      models.NotificationIngestCompleted,
				Status:    status,
				Message:   "Done",
				CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(defaultUsersMock()),
		argustest.WithAPINotifications(notifs),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	body := wire.NotificationUpdateRequest{Status: "read"}
	capture := rtesting.ServeRequest(engine, "PATCH", "/notifications/n-1", body)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.NotificationResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Status != models.NotificationRead {
		t.Errorf("Status: got %q, want %q", resp.Status, models.NotificationRead)
	}
}

func TestUpdateNotificationStatus_NotFound(t *testing.T) {
	notifs := &argustest.MockNotifications{
		OnUpdateStatus: func(_ context.Context, _, _, _ string, _ models.NotificationStatus) (*models.Notification, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(defaultUsersMock()),
		argustest.WithAPINotifications(notifs),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	body := wire.NotificationUpdateRequest{Status: "read"}
	capture := rtesting.ServeRequest(engine, "PATCH", "/notifications/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestBulkMarkRead_Success(t *testing.T) {
	var bulkCalled bool
	notifs := &argustest.MockNotifications{
		OnBulkUpdateStatus: func(_ context.Context, _, _ string, _ models.NotificationStatus) error {
			bulkCalled = true
			return nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(defaultUsersMock()),
		argustest.WithAPINotifications(notifs),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	body := wire.NotificationBulkReadRequest{Status: "read"}
	capture := rtesting.ServeRequest(engine, "POST", "/notifications/mark-read", body)
	rtesting.AssertStatus(t, capture, 200)
	if !bulkCalled {
		t.Error("expected BulkUpdateStatus to be called")
	}
}

func TestBulkMarkRead_UserNotFound(t *testing.T) {
	users := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithAPINotifications(&argustest.MockNotifications{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)
	body := wire.NotificationBulkReadRequest{Status: "read"}
	capture := rtesting.ServeRequest(engine, "POST", "/notifications/mark-read", body)
	rtesting.AssertStatus(t, capture, 404)
}
