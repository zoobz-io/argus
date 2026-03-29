//go:build testing

package handlers

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestNotificationSSE_Clone(t *testing.T) {
	orig := NotificationSSE{NotificationID: "n1", Type: "ingest.completed", Message: "hello"}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

func TestNotificationStream_Connected(t *testing.T) {
	users := defaultUsersMock()
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var capture *rtesting.StreamCapture
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		capture = rtesting.ServeStreamWithContext(ctx, engine, "GET", "/notifications/stream", nil)
	}()

	// Give handler time to subscribe and send connected.
	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	if len(evts) < 1 {
		t.Fatal("expected at least 1 event")
	}
	if evts[0].Event != "connected" {
		t.Errorf("expected event %q, got %q", "connected", evts[0].Event)
	}
}

func TestNotificationStream_ReceivesHint(t *testing.T) {
	users := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return &models.User{ID: "u-1", ExternalID: "user-1", TenantID: "tenant-1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var capture *rtesting.StreamCapture
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		capture = rtesting.ServeStreamWithContext(ctx, engine, "GET", "/notifications/stream", nil)
	}()

	time.Sleep(50 * time.Millisecond)

	// Emit a hint for the connected user.
	capitan.Emit(context.Background(), events.NotifyHintSignal, events.NotifyHintKey.Field(events.NotifyHint{
		UserID:         "u-1",
		TenantID:       "tenant-1",
		NotificationID: "n-1",
		Type:           "ingest.completed",
		Message:        "Document ready",
	}))

	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	if len(evts) < 2 {
		t.Fatalf("expected at least 2 events (connected + notification), got %d", len(evts))
	}

	found := false
	for _, evt := range evts {
		if evt.Event == "notification" {
			found = true
			var sse NotificationSSE
			if err := evt.DecodeJSON(&sse); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if sse.NotificationID != "n-1" {
				t.Errorf("NotificationID: got %q, want %q", sse.NotificationID, "n-1")
			}
		}
	}
	if !found {
		t.Error("expected a notification event")
	}
}

func TestNotificationStream_IgnoresOtherUsers(t *testing.T) {
	users := defaultUsersMock()
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var capture *rtesting.StreamCapture
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		capture = rtesting.ServeStreamWithContext(ctx, engine, "GET", "/notifications/stream", nil)
	}()

	time.Sleep(50 * time.Millisecond)

	// Emit a hint for a DIFFERENT user.
	capitan.Emit(context.Background(), events.NotifyHintSignal, events.NotifyHintKey.Field(events.NotifyHint{
		UserID:         "u-other",
		TenantID:       "tenant-1",
		NotificationID: "n-other",
		Type:           "ingest.completed",
		Message:        "Not for you",
	}))

	time.Sleep(50 * time.Millisecond)
	cancel()
	wg.Wait()

	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	for _, evt := range evts {
		if evt.Event == "notification" {
			t.Error("should not have received notification for different user")
		}
	}
}
