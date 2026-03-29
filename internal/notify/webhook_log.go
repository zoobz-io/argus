package notify

import (
	"context"

	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

func newWebhookLogStage() pipz.Chainable[*FanOutItem] {
	return pipz.Effect(
		WebhookLogID,
		func(ctx context.Context, item *FanOutItem) error {
			logger := sum.MustUse[intcontracts.NotifyDeliveryLogger](ctx)
			return logger.CreateDelivery(
				ctx,
				*item.Subscription.WebhookEndpointID,
				item.EventID,
				item.Notification.TenantID,
				item.WebhookStatusCode,
				1, // attempt number — backoff wrapper handles retries at the pipeline level
				item.WebhookDeliveryErr,
			)
		},
	)
}
