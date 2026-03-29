//go:build testing

package notify

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/sum"

	capitantesting "github.com/zoobz-io/capitan/testing"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
)

func setupPipelineRegistry(t *testing.T, indexer intcontracts.NotifyIndexer) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	sum.Register[intcontracts.NotifyIndexer](k, indexer)
	sum.Freeze(k)
	return context.Background()
}

func TestPipeline_FullIntegration(t *testing.T) {
	var indexedNotification *models.Notification
	indexer := &mockNotifyIndexer{
		OnIndex: func(_ context.Context, n *models.Notification) error {
			indexedNotification = n
			return nil
		},
	}

	ctx := setupPipelineRegistry(t, indexer)

	capture := capitantesting.NewEventCapture()
	listener := capitan.Hook(events.NotifyHintSignal, capture.Handler())
	defer listener.Close()

	pipeline := New()

	item := &FanOutItem{
		Notification: &models.Notification{
			TenantID: "t-1",
			Type:     models.NotificationIngestCompleted,
			Message:  "Document ingestion completed",
		},
		Subscription: &models.Subscription{
			ID:     "sub-1",
			UserID: "u-1",
		},
		EventID: "evt-1",
	}

	result, err := pipeline.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify assign stage
	if result.Notification.ID == "" {
		t.Error("expected notification ID to be assigned")
	}
	if result.Notification.UserID != "u-1" {
		t.Errorf("UserID: got %q, want %q", result.Notification.UserID, "u-1")
	}
	if result.Notification.EventID != "evt-1" {
		t.Errorf("EventID: got %q, want %q", result.Notification.EventID, "evt-1")
	}
	if result.Notification.Status != models.NotificationUnread {
		t.Errorf("Status: got %q, want %q", result.Notification.Status, models.NotificationUnread)
	}

	// Verify index stage
	if indexedNotification == nil {
		t.Fatal("expected notification to be indexed")
	}
	if indexedNotification.UserID != "u-1" {
		t.Errorf("indexed UserID: got %q, want %q", indexedNotification.UserID, "u-1")
	}

	// Verify hint stage
	if !capture.WaitForCount(1, time.Second) {
		t.Fatal("expected hint signal to be emitted")
	}
}

func TestPipeline_IndexError_PropagatesError(t *testing.T) {
	indexer := &mockNotifyIndexer{
		OnIndex: func(_ context.Context, _ *models.Notification) error {
			return errors.New("index unavailable")
		},
	}

	ctx := setupPipelineRegistry(t, indexer)
	pipeline := New()

	item := &FanOutItem{
		Notification: &models.Notification{
			TenantID: "t-1",
			Type:     models.NotificationIngestCompleted,
			Message:  "Document ingestion completed",
		},
		Subscription: &models.Subscription{
			ID:     "sub-1",
			UserID: "u-1",
		},
		EventID: "evt-1",
	}

	_, err := pipeline.Process(ctx, item)
	if err == nil {
		t.Fatal("expected error from index stage to propagate through pipeline")
	}
}
