//go:build testing

package notify

import (
	"context"
	"testing"
	"time"

	"github.com/zoobz-io/capitan"

	capitantesting "github.com/zoobz-io/capitan/testing"

	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
)

func TestHintStage_EmitsSignal(t *testing.T) {
	capture := capitantesting.NewEventCapture()
	listener := capitan.Hook(events.NotifyHintSignal, capture.Handler())
	defer listener.Close()

	stage := newHintStage()

	item := &FanOutItem{
		Notification: &models.Notification{
			ID:       "n-1",
			UserID:   "u-1",
			TenantID: "t-1",
			Type:     models.NotificationIngestCompleted,
			Message:  "Document ingestion completed",
		},
	}

	_, err := stage.Process(context.Background(), item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !capture.WaitForCount(1, time.Second) {
		t.Fatal("expected at least 1 captured event")
	}

	captured := capture.Events()
	if captured[0].Signal != events.NotifyHintSignal {
		t.Errorf("Signal: got %v, want %v", captured[0].Signal, events.NotifyHintSignal)
	}
}
