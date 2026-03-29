//go:build testing

package notify

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

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
