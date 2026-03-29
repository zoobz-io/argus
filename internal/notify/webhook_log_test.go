//go:build testing

package notify

import (
	"context"
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestWebhookLog_Success(t *testing.T) {
	var logged bool
	ctx := argustest.SetupRegistry(t, argustest.WithDeliveryLogger(&argustest.MockDeliveryLogger{
		OnCreateDelivery: func(_ context.Context, hookID, eventID, tenantID string, statusCode, attempt int, _ *string) error {
			if hookID != "h-1" {
				t.Errorf("hookID = %s, want h-1", hookID)
			}
			if eventID != "evt-1" {
				t.Errorf("eventID = %s, want evt-1", eventID)
			}
			if tenantID != "t-1" {
				t.Errorf("tenantID = %s, want t-1", tenantID)
			}
			if statusCode != 200 {
				t.Errorf("statusCode = %d, want 200", statusCode)
			}
			if attempt != 1 {
				t.Errorf("attempt = %d, want 1", attempt)
			}
			logged = true
			return nil
		},
	}))

	endpointID := "h-1"
	item := &FanOutItem{
		Notification:      &models.Notification{TenantID: "t-1"},
		Subscription:      &models.Subscription{WebhookEndpointID: &endpointID},
		EventID:           "evt-1",
		WebhookStatusCode: 200,
	}

	stage := newWebhookLogStage()
	_, err := stage.Process(ctx, item)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !logged {
		t.Error("delivery was not logged")
	}
}
