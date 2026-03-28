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

func newTestWatchedPaths(t *testing.T, mock *soytesting.MockDB) *WatchedPaths {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewWatchedPaths(mock.DB(), astqlpg.New())
}

func TestWatchedPaths_GetWatchedPath(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-1", TenantID: "t-1", ProviderID: "p-1", Path: "/docs", Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	wp, err := store.GetWatchedPath(context.Background(), "wp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wp.ID != "wp-1" {
		t.Errorf("ID: got %q, want %q", wp.ID, "wp-1")
	}
	if wp.Path != "/docs" {
		t.Errorf("Path: got %q, want %q", wp.Path, "/docs")
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_CreateWatchedPath(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-new", TenantID: "t-1", ProviderID: "p-1", Path: "/docs", Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	wp, err := store.CreateWatchedPath(context.Background(), "t-1", "p-1", "/docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wp.Path != "/docs" {
		t.Errorf("Path: got %q, want %q", wp.Path, "/docs")
	}
	if wp.Active != true {
		t.Errorf("Active: got %v, want true", wp.Active)
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_CreateWatchedPath_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate path"))

	_, err := store.CreateWatchedPath(context.Background(), "t-1", "p-1", "/docs")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		if got := err.Error(); got == "" {
			t.Error("expected non-empty error message")
		}
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_UpdateWatchedPath(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// GetWatchedPath (SELECT)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-1", TenantID: "t-1", ProviderID: "p-1", Path: "/old", Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-1", TenantID: "t-1", ProviderID: "p-1", Path: "/new", Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	wp, err := store.UpdateWatchedPath(context.Background(), "wp-1", "/new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if wp.Path != "/new" {
		t.Errorf("Path: got %q, want %q", wp.Path, "/new")
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_DeleteWatchedPath(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteWatchedPath(context.Background(), "wp-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_ListWatchedPaths(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-1", TenantID: "t-1", ProviderID: "p-1", Path: "/docs", Active: true, CreatedAt: ts, UpdatedAt: ts},
		{ID: "wp-2", TenantID: "t-1", ProviderID: "p-1", Path: "/images", Active: true, CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})

	result, err := store.ListWatchedPaths(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_ListWatchedPathsByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-1", TenantID: "t-1", ProviderID: "p-1", Path: "/docs", Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	result, err := store.ListWatchedPathsByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_ListWatchedPaths_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListWatchedPaths(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestWatchedPathsByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListWatchedPathsByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_UpdateWatchedPath_GetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.UpdateWatchedPath(context.Background(), "wp-1", "/new")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestWatchedPaths_UpdateWatchedPath_SetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestWatchedPaths(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.WatchedPath{
		{ID: "wp-1", TenantID: "t-1", ProviderID: "p-1", Path: "/old", Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.UpdateWatchedPath(context.Background(), "wp-1", "/new")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
