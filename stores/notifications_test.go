//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

func TestNotifications_Index(t *testing.T) {
	var indexedID string
	var indexedData []byte
	mock := &mockSearchProvider{
		OnIndex: func(_ context.Context, _, id string, doc []byte) error {
			indexedID = id
			indexedData = doc
			return nil
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)

	n := &models.Notification{
		ID:       "n-1",
		TenantID: "t-1",
		Type:     models.NotificationIngestCompleted,
		Status:   models.NotificationUnread,
		Message:  "Document ingestion completed",
	}

	err := store.Index(context.Background(), n)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if indexedID != "n-1" {
		t.Errorf("indexed ID: got %q, want %q", indexedID, "n-1")
	}
	if len(indexedData) == 0 {
		t.Error("expected non-empty indexed data")
	}
}

func TestNotifications_Index_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnIndex: func(_ context.Context, _, _ string, _ []byte) error {
			return errors.New("index failure")
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)

	n := &models.Notification{
		ID:        "n-1",
		CreatedAt: time.Now(),
	}

	err := store.Index(context.Background(), n)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotifications_SearchByUser(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Total: 1,
				Hits: []grub.SearchHit{
					{Source: jsonBytes(t, models.Notification{
						ID: "n-1", UserID: "u-1", TenantID: "t-1",
						Type: models.NotificationIngestCompleted, Status: models.NotificationUnread,
						Message: "done", CreatedAt: ts,
					})},
				},
			}, nil
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	result, err := store.SearchByUser(context.Background(), "t-1", "u-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total: got %d, want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(result.Items))
	}
	if result.Items[0].ID != "n-1" {
		t.Errorf("ID: got %q, want %q", result.Items[0].ID, "n-1")
	}
}

func TestNotifications_SearchByUser_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return nil, errors.New("search failure")
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	_, err := store.SearchByUser(context.Background(), "t-1", "u-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotifications_UpdateStatus(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnGet: func(_ context.Context, _, _ string) ([]byte, error) {
			return jsonBytes(t, models.Notification{
				ID: "n-1", UserID: "u-1", TenantID: "t-1",
				Type: models.NotificationIngestCompleted, Status: models.NotificationUnread,
				Message: "done", CreatedAt: ts,
			}), nil
		},
		OnIndex: func(_ context.Context, _, _ string, _ []byte) error {
			return nil
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	n, err := store.UpdateStatus(context.Background(), "t-1", "u-1", "n-1", models.NotificationRead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.Status != models.NotificationRead {
		t.Errorf("Status: got %q, want %q", n.Status, models.NotificationRead)
	}
}

func TestNotifications_UpdateStatus_NotFound(t *testing.T) {
	mock := &mockSearchProvider{
		OnGet: func(_ context.Context, _, _ string) ([]byte, error) {
			return nil, errors.New("not found")
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	_, err := store.UpdateStatus(context.Background(), "t-1", "u-1", "n-1", models.NotificationRead)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotifications_UpdateStatus_WrongOwner(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnGet: func(_ context.Context, _, _ string) ([]byte, error) {
			return jsonBytes(t, models.Notification{
				ID: "n-1", UserID: "u-other", TenantID: "t-1",
				Type: models.NotificationIngestCompleted, Status: models.NotificationUnread,
				Message: "done", CreatedAt: ts,
			}), nil
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	_, err := store.UpdateStatus(context.Background(), "t-1", "u-1", "n-1", models.NotificationRead)
	if err == nil {
		t.Fatal("expected error for wrong owner")
	}
}

func TestNotifications_UpdateStatus_IndexError(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnGet: func(_ context.Context, _, _ string) ([]byte, error) {
			return jsonBytes(t, models.Notification{
				ID: "n-1", UserID: "u-1", TenantID: "t-1",
				Type: models.NotificationIngestCompleted, Status: models.NotificationUnread,
				Message: "done", CreatedAt: ts,
			}), nil
		},
		OnIndex: func(_ context.Context, _, _ string, _ []byte) error {
			return errors.New("index failure")
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	_, err := store.UpdateStatus(context.Background(), "t-1", "u-1", "n-1", models.NotificationRead)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotifications_BulkUpdateStatus(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	var batchCalled bool
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Total: 2,
				Hits: []grub.SearchHit{
					{Source: jsonBytes(t, models.Notification{
						ID: "n-1", UserID: "u-1", TenantID: "t-1",
						Type: models.NotificationIngestCompleted, Status: models.NotificationUnread,
						Message: "done1", CreatedAt: ts,
					})},
					{Source: jsonBytes(t, models.Notification{
						ID: "n-2", UserID: "u-1", TenantID: "t-1",
						Type: models.NotificationIngestFailed, Status: models.NotificationUnread,
						Message: "done2", CreatedAt: ts,
					})},
				},
			}, nil
		},
	}
	// Override IndexBatch to track the call.
	mock.OnIndexBatch = func(_ context.Context, _ string, docs map[string][]byte) error {
		batchCalled = true
		if len(docs) != 2 {
			t.Errorf("expected 2 docs in batch, got %d", len(docs))
		}
		return nil
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	err := store.BulkUpdateStatus(context.Background(), "t-1", "u-1", models.NotificationRead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !batchCalled {
		t.Error("expected IndexBatch to be called")
	}
}

func TestNotifications_BulkUpdateStatus_Empty(t *testing.T) {
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{Total: 0, Hits: []grub.SearchHit{}}, nil
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	err := store.BulkUpdateStatus(context.Background(), "t-1", "u-1", models.NotificationRead)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestNotifications_BulkUpdateStatus_SearchError(t *testing.T) {
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return nil, errors.New("search failure")
		},
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	err := store.BulkUpdateStatus(context.Background(), "t-1", "u-1", models.NotificationRead)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNotifications_BulkUpdateStatus_BatchError(t *testing.T) {
	ts := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Total: 1,
				Hits: []grub.SearchHit{
					{Source: jsonBytes(t, models.Notification{
						ID: "n-1", UserID: "u-1", TenantID: "t-1",
						Status: models.NotificationUnread, CreatedAt: ts,
					})},
				},
			}, nil
		},
	}
	mock.OnIndexBatch = func(_ context.Context, _ string, _ map[string][]byte) error {
		return errors.New("batch failure")
	}

	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()

	store := NewNotifications(mock)
	err := store.BulkUpdateStatus(context.Background(), "t-1", "u-1", models.NotificationRead)
	if err == nil {
		t.Fatal("expected error")
	}
}
