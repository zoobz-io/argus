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

func newTestDeliveries(t *testing.T, mock *soytesting.MockDB) *Deliveries {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	k := sum.Start()
	sum.Freeze(k)
	return NewDeliveries(mock.DB(), astqlpg.New())
}

func TestDeliveries_CreateDelivery(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Delivery{
		{ID: "d-new", HookID: "h-1", EventID: "evt-1", TenantID: "t-1", StatusCode: 200, Attempt: 1, CreatedAt: ts},
	})

	err := store.CreateDelivery(context.Background(), "h-1", "evt-1", "t-1", 200, 1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestDeliveries_CreateDelivery_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	mock.ExpectExec().WithError(errors.New("db error"))

	err := store.CreateDelivery(context.Background(), "h-1", "evt-1", "t-1", 500, 1, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDeliveries_CreateDelivery_WithError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	errMsg := "timeout"
	mock.ExpectQuery().WithRows([]models.Delivery{
		{ID: "d-new", HookID: "h-1", EventID: "evt-1", TenantID: "t-1", StatusCode: 0, Attempt: 1, Error: &errMsg, CreatedAt: ts},
	})

	err := store.CreateDelivery(context.Background(), "h-1", "evt-1", "t-1", 0, 1, &errMsg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestDeliveries_ListByHook(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Delivery{
		{ID: "d-1", HookID: "h-1", EventID: "evt-1", TenantID: "t-1", StatusCode: 200, Attempt: 1, CreatedAt: ts},
		{ID: "d-2", HookID: "h-1", EventID: "evt-2", TenantID: "t-1", StatusCode: 500, Attempt: 1, CreatedAt: ts.Add(time.Hour)},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListByHook(context.Background(), "t-1", "h-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	if result.Total != 5 {
		t.Errorf("Total: got %d, want 5", result.Total)
	}
	mock.AssertExpectations()
}

func TestDeliveries_ListByHook_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListByHook(context.Background(), "t-1", "h-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDeliveries_ListByHook_CountError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Delivery{
		{ID: "d-1", HookID: "h-1", EventID: "evt-1", TenantID: "t-1", StatusCode: 200, Attempt: 1, CreatedAt: ts},
	})
	mock.ExpectQuery().WithError(errors.New("count error"))

	_, err := store.ListByHook(context.Background(), "t-1", "h-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error from count query")
	}
	mock.AssertExpectations()
}

func TestDeliveries_ListDeliveries(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Delivery{
		{ID: "d-1", HookID: "h-1", EventID: "evt-1", TenantID: "t-1", StatusCode: 200, Attempt: 1, CreatedAt: ts},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 10}})

	result, err := store.ListDeliveries(context.Background(), models.OffsetPage{Offset: 0, Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	if result.Total != 10 {
		t.Errorf("Total: got %d, want 10", result.Total)
	}
	mock.AssertExpectations()
}

func TestDeliveries_ListDeliveries_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDeliveries(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListDeliveries(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
