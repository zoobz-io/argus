package notify

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zoobz-io/pipz"
)

var webhookHTTPClient = &http.Client{Timeout: 10 * time.Second}

func newWebhookDeliverStage() pipz.Chainable[*FanOutItem] {
	return pipz.Apply(
		WebhookDeliverID,
		func(ctx context.Context, item *FanOutItem) (*FanOutItem, error) {
			if item.WebhookHook == nil {
				return item, fmt.Errorf("webhook hook not loaded by sign stage")
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, item.WebhookHook.URL, bytes.NewReader(item.WebhookPayload))
			if err != nil {
				return item, fmt.Errorf("creating request: %w", err)
			}

			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Argus-Signature", item.WebhookSignature)
			eventType := string(item.Notification.Type)
			if item.DomainEvent != nil {
				eventType = item.DomainEvent.Action
			}
			req.Header.Set("X-Argus-Event", eventType)
			req.Header.Set("X-Argus-Delivery", item.Notification.ID)
			req.Header.Set("X-Argus-Timestamp", item.WebhookTimestamp)

			resp, err := webhookHTTPClient.Do(req)
			if err != nil {
				errMsg := err.Error()
				item.WebhookDeliveryErr = &errMsg
				return item, fmt.Errorf("delivering webhook: %w", err)
			}
			defer func() { _ = resp.Body.Close() }()

			item.WebhookStatusCode = resp.StatusCode
			if resp.StatusCode >= 400 {
				errMsg := fmt.Sprintf("endpoint returned %d", resp.StatusCode)
				item.WebhookDeliveryErr = &errMsg
				return item, fmt.Errorf("webhook delivery failed: %s", errMsg)
			}
			return item, nil
		},
	)
}
