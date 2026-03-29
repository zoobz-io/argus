//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func defaultUsersMock() *argustest.MockUsers {
	return &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return &models.User{ID: "u-1", ExternalID: "user-1", TenantID: "tenant-1"}, nil
		},
	}
}

func TestListMySubscriptions_Success(t *testing.T) {
	users := defaultUsersMock()
	subs := &argustest.MockSubscriptions{
		OnListSubscriptionsByUser: func(_ context.Context, tenantID, userID string, _ models.OffsetPage) (*models.OffsetResult[models.Subscription], error) {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %q", tenantID)
			}
			if userID != "u-1" {
				t.Errorf("expected u-1, got %q", userID)
			}
			return &models.OffsetResult[models.Subscription]{
				Items: []*models.Subscription{{ID: "sub-1", UserID: "u-1", EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox}},
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(users), argustest.WithAPISubscriptions(subs), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/subscriptions", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.SubscriptionListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Subscriptions) != 1 || resp.Subscriptions[0].ID != "sub-1" {
		t.Errorf("unexpected subscriptions: %+v", resp.Subscriptions)
	}
}

func TestListMySubscriptions_UserNotFound(t *testing.T) {
	users := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(users), argustest.WithAPISubscriptions(&argustest.MockSubscriptions{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/subscriptions", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateSubscription_Success(t *testing.T) {
	users := defaultUsersMock()
	subs := &argustest.MockSubscriptions{
		OnCreateSubscription: func(_ context.Context, tenantID, userID, eventType string, channel models.SubscriptionChannel, _ string) (*models.Subscription, error) {
			return &models.Subscription{ID: "sub-new", UserID: userID, TenantID: tenantID, EventType: eventType, Channel: channel}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(users), argustest.WithAPISubscriptions(subs), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.SubscriptionRequest{EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox}
	capture := rtesting.ServeRequest(engine, "POST", "/subscriptions", body)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.SubscriptionResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "sub-new" {
		t.Errorf("unexpected ID: %q", resp.ID)
	}
}

func TestDeleteSubscription_Success(t *testing.T) {
	subs := &argustest.MockSubscriptions{
		OnDeleteSubscription: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(defaultUsersMock()), argustest.WithAPISubscriptions(subs), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "DELETE", "/subscriptions/sub-1", nil)
	if capture.StatusCode() >= 400 {
		t.Errorf("expected success status, got %d", capture.StatusCode())
	}
}

func TestDeleteSubscription_NotFound(t *testing.T) {
	subs := &argustest.MockSubscriptions{
		OnDeleteSubscription: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(defaultUsersMock()), argustest.WithAPISubscriptions(subs), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "DELETE", "/subscriptions/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
