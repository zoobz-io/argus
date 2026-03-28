//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

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
