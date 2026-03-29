//go:build testing

package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestWebhookDeliver_Success(t *testing.T) {
	received := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Argus-Signature") == "" {
			t.Error("missing X-Argus-Signature header")
		}
		if r.Header.Get("X-Argus-Event") == "" {
			t.Error("missing X-Argus-Event header")
		}
		if r.Header.Get("X-Argus-Timestamp") == "" {
			t.Error("missing X-Argus-Timestamp header")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type = %s, want application/json", r.Header.Get("Content-Type"))
		}
		w.WriteHeader(200)
		close(received)
	}))
	defer srv.Close()

	ctx := argustest.SetupRegistry(t)
	item := &FanOutItem{
		Notification:     &models.Notification{ID: "n-1", Type: models.NotificationIngestCompleted},
		WebhookHook:      &models.Hook{URL: srv.URL},
		WebhookPayload:   []byte(`{"id":"n-1"}`),
		WebhookSignature: "sha256=abc",
		WebhookTimestamp: "1234567890",
	}

	stage := newWebhookDeliverStage()
	result, err := stage.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.WebhookStatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.WebhookStatusCode)
	}
	<-received
}

func TestWebhookDeliver_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(500)
	}))
	defer srv.Close()

	ctx := argustest.SetupRegistry(t)
	item := &FanOutItem{
		Notification:     &models.Notification{ID: "n-1", Type: models.NotificationIngestCompleted},
		WebhookHook:      &models.Hook{URL: srv.URL},
		WebhookPayload:   []byte(`{}`),
		WebhookSignature: "sha256=abc",
		WebhookTimestamp: "1234567890",
	}

	stage := newWebhookDeliverStage()
	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Error("expected error for 500 response")
	}
	// The item is mutated in place before the error is returned.
	if item.WebhookStatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", item.WebhookStatusCode)
	}
	if item.WebhookDeliveryErr == nil {
		t.Error("expected delivery error to be set")
	}
}

func TestWebhookDeliver_NoHook(t *testing.T) {
	ctx := context.Background()
	item := &FanOutItem{
		Notification: &models.Notification{},
	}
	stage := newWebhookDeliverStage()
	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Error("expected error for nil hook")
	}
}

func TestWebhookDeliver_ConnectionRefused(t *testing.T) {
	// Use a URL that will definitely refuse connections.
	ctx := argustest.SetupRegistry(t)
	item := &FanOutItem{
		Notification:     &models.Notification{ID: "n-1", Type: models.NotificationIngestCompleted},
		WebhookHook:      &models.Hook{URL: "http://127.0.0.1:1"},
		WebhookPayload:   []byte(`{}`),
		WebhookSignature: "sha256=abc",
		WebhookTimestamp: "1234567890",
	}

	stage := newWebhookDeliverStage()
	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Fatal("expected error for connection refused")
	}
	if item.WebhookDeliveryErr == nil {
		t.Error("expected delivery error to be set")
	}
}

func TestWebhookDeliver_Non2xxResponses(t *testing.T) {
	codes := []int{400, 403, 404, 502, 503}
	for _, code := range codes {
		t.Run(http.StatusText(code), func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(code)
			}))
			defer srv.Close()

			ctx := argustest.SetupRegistry(t)
			item := &FanOutItem{
				Notification:     &models.Notification{ID: "n-1", Type: models.NotificationIngestCompleted},
				WebhookHook:      &models.Hook{URL: srv.URL},
				WebhookPayload:   []byte(`{}`),
				WebhookSignature: "sha256=abc",
				WebhookTimestamp: "1234567890",
			}

			stage := newWebhookDeliverStage()
			_, err := stage.Process(ctx, item)
			if err == nil {
				t.Errorf("expected error for %d response", code)
			}
			if item.WebhookStatusCode != code {
				t.Errorf("StatusCode = %d, want %d", item.WebhookStatusCode, code)
			}
			if item.WebhookDeliveryErr == nil {
				t.Error("expected delivery error to be set")
			}
		})
	}
}

func TestWebhookDeliver_Timeout(t *testing.T) {
	// Create a server that delays long enough for the context to expire.
	done := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		select {
		case <-done:
		case <-time.After(10 * time.Second):
		}
	}))
	defer func() {
		close(done)
		srv.Close()
	}()

	// Use a context with a short timeout.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	item := &FanOutItem{
		Notification:     &models.Notification{ID: "n-1", Type: models.NotificationIngestCompleted},
		WebhookHook:      &models.Hook{URL: srv.URL},
		WebhookPayload:   []byte(`{}`),
		WebhookSignature: "sha256=abc",
		WebhookTimestamp: "1234567890",
	}

	stage := newWebhookDeliverStage()
	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Fatal("expected error for timeout")
	}
	if item.WebhookDeliveryErr == nil {
		t.Error("expected delivery error to be set")
	}
}
