//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"

	soytesting "github.com/zoobz-io/soy/testing"
)

func newTestSubscriptions(t *testing.T, mock *soytesting.MockDB) *Subscriptions {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewSubscriptions(mock.DB(), astqlpg.New())
}

func testSubscription() models.Subscription {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return models.Subscription{
		ID:        "sub-1",
		UserID:    "u-1",
		TenantID:  "t-1",
		EventType: "ingest.completed",
		Channel:   models.SubscriptionChannelInbox,
		CreatedAt: ts,
		UpdatedAt: ts,
	}
}

func TestSubscriptions_CreateSubscription(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s := testSubscription()
	mock.ExpectQuery().WithRows([]models.Subscription{s})

	sub, err := store.CreateSubscription(context.Background(), "t-1", "u-1", "ingest.completed", models.SubscriptionChannelInbox)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.EventType != "ingest.completed" {
		t.Errorf("EventType: got %q, want %q", sub.EventType, "ingest.completed")
	}
	if sub.Channel != models.SubscriptionChannelInbox {
		t.Errorf("Channel: got %q, want %q", sub.Channel, models.SubscriptionChannelInbox)
	}
	mock.AssertExpectations()
}

func TestSubscriptions_CreateSubscription_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate"))

	_, err := store.CreateSubscription(context.Background(), "t-1", "u-1", "ingest.completed", models.SubscriptionChannelInbox)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_GetSubscription(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s := testSubscription()
	mock.ExpectQuery().WithRows([]models.Subscription{s})

	sub, err := store.GetSubscription(context.Background(), "sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("ID: got %q, want %q", sub.ID, "sub-1")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_GetSubscription_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetSubscription(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_GetSubscriptionByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s := testSubscription()
	mock.ExpectQuery().WithRows([]models.Subscription{s})

	sub, err := store.GetSubscriptionByTenant(context.Background(), "t-1", "sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("ID: got %q, want %q", sub.ID, "sub-1")
	}
	if sub.TenantID != "t-1" {
		t.Errorf("TenantID: got %q, want %q", sub.TenantID, "t-1")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_GetSubscriptionByTenant_NotFound(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectQuery().WithRows([]models.Subscription{})

	_, err := store.GetSubscriptionByTenant(context.Background(), "wrong-tenant", "sub-1")
	if err == nil {
		t.Fatal("expected error for tenant mismatch")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_GetSubscriptionByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.GetSubscriptionByTenant(context.Background(), "t-1", "sub-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_DeleteSubscription(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s := testSubscription()
	mock.ExpectQuery().WithRows([]models.Subscription{s}) // GetSubscriptionByTenant
	mock.ExpectExec().WithResult(1, 0)                    // Delete

	err := store.DeleteSubscription(context.Background(), "t-1", "u-1", "sub-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestSubscriptions_ListSubscriptionsByUser(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s1 := testSubscription()
	s2 := testSubscription()
	s2.ID = "sub-2"
	s2.EventType = "ingest.failed"
	mock.ExpectQuery().WithRows([]models.Subscription{s1, s2})

	result, err := store.ListSubscriptionsByUser(context.Background(), "t-1", "u-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestSubscriptions_ListSubscriptionsByUser_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListSubscriptionsByUser(context.Background(), "t-1", "u-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_FindByTenantAndEventType(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s := testSubscription()
	mock.ExpectQuery().WithRows([]models.Subscription{s})

	subs, err := store.FindByTenantAndEventType(context.Background(), "t-1", "ingest.completed")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 1 {
		t.Errorf("got %d subscriptions, want 1", len(subs))
	}
	mock.AssertExpectations()
}

func TestSubscriptions_FindByTenantAndEventType_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.FindByTenantAndEventType(context.Background(), "t-1", "ingest.completed")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestSubscriptions_CreateDefaultSubscriptions(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	s := testSubscription()
	// Two inserts (ingest.completed + ingest.failed), each returns a query with rows.
	mock.ExpectQuery().WithRows([]models.Subscription{s})
	s2 := testSubscription()
	s2.EventType = "ingest.failed"
	mock.ExpectQuery().WithRows([]models.Subscription{s2})

	err := store.CreateDefaultSubscriptions(context.Background(), "u-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestSubscriptions_CreateDefaultSubscriptions_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestSubscriptions(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	err := store.CreateDefaultSubscriptions(context.Background(), "u-1", "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
