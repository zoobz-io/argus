//go:build testing

package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
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

	// Verify payload is valid JSON with notification context.
	var p struct {
		Notification *webhookNotif `json:"notification"`
	}
	if err := json.Unmarshal(result.WebhookPayload, &p); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if p.Notification == nil || p.Notification.ID != "n-1" {
		id := ""
		if p.Notification != nil {
			id = p.Notification.ID
		}
		t.Errorf("payload notification ID = %s, want n-1", id)
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

func TestWebhookSign_EmptySecret(t *testing.T) {
	// A hook with an empty secret should still produce a valid HMAC (with empty key).
	hook := &models.Hook{ID: "h-1", TenantID: "t-1", Secret: "", URL: "https://example.com/hook"}
	ctx := argustest.SetupRegistry(t, argustest.WithHookLoader(&argustest.MockHookLoader{
		OnGetWithSecret: func(_ context.Context, _, _ string) (*models.Hook, error) {
			return hook, nil
		},
	}))

	endpointID := "h-1"
	item := &FanOutItem{
		Notification: &models.Notification{ID: "n-1", TenantID: "t-1", Type: models.NotificationIngestCompleted},
		Subscription: &models.Subscription{WebhookEndpointID: &endpointID},
		EventID:      "evt-1",
	}

	stage := newWebhookSignStage()
	result, err := stage.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// HMAC with empty key still produces a signature.
	if result.WebhookSignature == "" {
		t.Error("expected non-empty signature even with empty secret")
	}
	if !strings.HasPrefix(result.WebhookSignature, "sha256=") {
		t.Errorf("signature should start with sha256=, got %q", result.WebhookSignature)
	}

	// Verify the HMAC is correct with empty key.
	mac := hmac.New(sha256.New, []byte(""))
	mac.Write([]byte(result.WebhookTimestamp + "."))
	mac.Write(result.WebhookPayload)
	expected := "sha256=" + hex.EncodeToString(mac.Sum(nil))
	if result.WebhookSignature != expected {
		t.Errorf("signature mismatch with empty key: got %s, want %s", result.WebhookSignature, expected)
	}
}

func TestWebhookSign_HookSetsFields(t *testing.T) {
	// Verify that the sign stage correctly populates all webhook fields on the item.
	hook := &models.Hook{ID: "h-2", TenantID: "t-1", Secret: "another-secret", URL: "https://example.com/webhook"}
	ctx := argustest.SetupRegistry(t, argustest.WithHookLoader(&argustest.MockHookLoader{
		OnGetWithSecret: func(_ context.Context, _, _ string) (*models.Hook, error) {
			return hook, nil
		},
	}))

	endpointID := "h-2"
	item := &FanOutItem{
		Notification: &models.Notification{ID: "n-2", TenantID: "t-1", Type: models.NotificationIngestCompleted, Message: "test"},
		Subscription: &models.Subscription{WebhookEndpointID: &endpointID},
		EventID:      "evt-2",
	}

	stage := newWebhookSignStage()
	result, err := stage.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.WebhookHook.ID != "h-2" {
		t.Errorf("hook ID = %s, want h-2", result.WebhookHook.ID)
	}
	if result.WebhookHook.URL != "https://example.com/webhook" {
		t.Errorf("hook URL = %s", result.WebhookHook.URL)
	}

	// Payload should be valid JSON with notification context.
	var p struct {
		Notification *webhookNotif `json:"notification"`
	}
	if err := json.Unmarshal(result.WebhookPayload, &p); err != nil {
		t.Fatalf("payload is not valid JSON: %v", err)
	}
	if p.Notification == nil || p.Notification.ID != "n-2" {
		id := ""
		if p.Notification != nil {
			id = p.Notification.ID
		}
		t.Errorf("payload notification ID = %s, want n-2", id)
	}
}
