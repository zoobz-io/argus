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

			payload, err := json.Marshal(item.Notification)
			if err != nil {
				return item, fmt.Errorf("marshaling payload: %w", err)
			}

			timestamp := strconv.FormatInt(time.Now().Unix(), 10)

			mac := hmac.New(sha256.New, []byte(hook.Secret))
			mac.Write([]byte(timestamp))
			mac.Write([]byte("."))
			mac.Write(payload)
			signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

			item.WebhookHook = hook
			item.WebhookTimestamp = timestamp
			item.WebhookPayload = payload
			item.WebhookSignature = signature
			return item, nil
		},
	)
}
