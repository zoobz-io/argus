//go:build testing

package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestWebhookSign_Success(t *testing.T) {
	hook := &models.Hook{ID: "h-1", TenantID: "t-1", Secret: "test-secret", URL: "https://example.com/hook"}
	ctx := argustest.SetupRegistry(t, argustest.WithHookLoader(&argustest.MockHookLoader{
		OnGetWithSecret: func(_ context.Context, _, _ string) (*models.Hook, error) {
			return hook, nil
		},
	}))

	endpointID := "h-1"
	item := &FanOutItem{
		Notification: &models.Notification{ID: "n-1", TenantID: "t-1", Type: models.NotificationIngestCompleted, Message: "done"},
		Subscription: &models.Subscription{WebhookEndpointID: &endpointID},
		EventID:      "evt-1",
	}

	stage := newWebhookSignStage()
	result, err := stage.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.WebhookHook == nil {
		t.Fatal("WebhookHook not set on item")
	}
	if len(result.WebhookPayload) == 0 {
		t.Fatal("WebhookPayload is empty")
	}
	if result.WebhookSignature == "" {
		t.Fatal("WebhookSignature is empty")
	}
	if result.WebhookTimestamp == "" {
		t.Fatal("WebhookTimestamp is empty")
	}

	// Verify HMAC independently.
	mac := hmac.New(sha256.New, []byte("test-secret"))
	mac.Write([]byte(result.WebhookTimestamp + "."))
	mac.Write(result.WebhookPayload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if result.WebhookSignature != expected {
		t.Errorf("signature mismatch: got %s, want %s", result.WebhookSignature, expected)
	}

	// Verify payload is valid JSON of the notification.
	var n models.Notification
	if err := json.Unmarshal(result.WebhookPayload, &n); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if n.ID != "n-1" {
		t.Errorf("payload notification ID = %s, want n-1", n.ID)
	}
}

func TestWebhookSign_NoEndpointID(t *testing.T) {
	ctx := argustest.SetupRegistry(t)
	item := &FanOutItem{
		Notification: &models.Notification{},
		Subscription: &models.Subscription{},
	}
	stage := newWebhookSignStage()
	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Error("expected error for nil webhook endpoint ID")
	}
}

func TestWebhookSign_LoaderError(t *testing.T) {
	ctx := argustest.SetupRegistry(t, argustest.WithHookLoader(&argustest.MockHookLoader{
		OnGetWithSecret: func(_ context.Context, _, _ string) (*models.Hook, error) {
			return nil, sum.ErrNotFound
		},
	}))

	endpointID := "missing"
	item := &FanOutItem{
		Notification: &models.Notification{TenantID: "t-1"},
		Subscription: &models.Subscription{WebhookEndpointID: &endpointID},
	}
	stage := newWebhookSignStage()
	_, err := stage.Process(ctx, item)
	if err == nil {
		t.Error("expected error for loader failure")
	}
}
