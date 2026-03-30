package notify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

func newWebhookSignStage() pipz.Chainable[*FanOutItem] {
	return pipz.Apply(
		WebhookSignID,
		func(ctx context.Context, item *FanOutItem) (*FanOutItem, error) {
			if item.Subscription.WebhookEndpointID == nil {
				return item, fmt.Errorf("subscription has no webhook endpoint")
			}

			loader := sum.MustUse[intcontracts.NotifyHookLoader](ctx)
			hook, err := loader.GetWithSecret(ctx, item.Notification.TenantID, *item.Subscription.WebhookEndpointID)
			if err != nil {
				return item, fmt.Errorf("loading hook: %w", err)
			}

			payloadData := buildWebhookPayload(item)
			payload, err := json.Marshal(payloadData)
			if err != nil {
				return item, fmt.Errorf("marshaling payload: %w", err)
			}

			timestamp := strconv.FormatInt(time.Now().Unix(), 10)

			// Include timestamp in signed content to bind signature to delivery time.
			mac := hmac.New(sha256.New, []byte(hook.Secret))
			mac.Write([]byte(timestamp + "."))
			mac.Write(payload)
			signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

			item.WebhookHook = hook
			item.WebhookPayload = payload
			item.WebhookSignature = signature
			item.WebhookTimestamp = timestamp
			return item, nil
		},
	)
}

// webhookPayload mirrors wire.WebhookPayload without importing api/wire.
//
//nolint:govet // fieldalignment: mirrors wire type layout
type webhookPayload struct {
	Timestamp    any             `json:"timestamp"`
	Action       string          `json:"action"`
	ResourceType string          `json:"resource_type"`
	ResourceID   string          `json:"resource_id"`
	TenantID     string          `json:"tenant_id"`
	ActorID      string          `json:"actor_id"`
	Metadata     json.RawMessage `json:"metadata,omitempty"`
	Notification *webhookNotif   `json:"notification"`
}

type webhookNotif struct {
	ID      string `json:"id"`
	UserID  string `json:"user_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func buildWebhookPayload(item *FanOutItem) *webhookPayload {
	p := &webhookPayload{
		Notification: &webhookNotif{
			ID:      item.Notification.ID,
			UserID:  item.Notification.UserID,
			Status:  string(item.Notification.Status),
			Message: item.Notification.Message,
		},
	}
	if item.DomainEvent != nil {
		p.Timestamp = item.DomainEvent.Timestamp
		p.Action = item.DomainEvent.Action
		p.ResourceType = item.DomainEvent.ResourceType
		p.ResourceID = item.DomainEvent.ResourceID
		p.TenantID = item.DomainEvent.TenantID
		p.ActorID = item.DomainEvent.ActorID
		p.Metadata = item.DomainEvent.Metadata
	} else {
		p.Action = string(item.Notification.Type)
		p.TenantID = item.Notification.TenantID
		p.Metadata = item.Notification.Metadata
	}
	return p
}
