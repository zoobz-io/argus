package notify

import (
	"context"
	"time"

	"github.com/zoobz-io/pipz"
)

// Webhook pipeline stage identities.
var (
	WebhookPipelineID = pipz.NewIdentity("webhook-pipeline", "Webhook delivery pipeline")
	WebhookSignID     = pipz.NewIdentity("webhook-sign", "Sign webhook payload with HMAC-SHA256")
	WebhookDeliverID  = pipz.NewIdentity("webhook-deliver", "Deliver webhook payload via HTTP POST")
	WebhookLogID      = pipz.NewIdentity("webhook-log", "Log webhook delivery attempt")
)

// WebhookPipeline orchestrates webhook delivery through signing, delivery,
// and logging stages. The deliver stage is wrapped in backoff for retry.
type WebhookPipeline struct {
	sequence *pipz.Sequence[*FanOutItem]
}

// NewWebhookPipeline creates a new webhook delivery pipeline.
func NewWebhookPipeline() *WebhookPipeline {
	deliverWithRetry := pipz.NewBackoff(
		pipz.NewIdentity("webhook-deliver-retry", "Retry webhook delivery with backoff"),
		newWebhookDeliverStage(),
		3,
		time.Second,
	)

	seq := pipz.NewSequence(
		WebhookPipelineID,
		newWebhookSignStage(),
		deliverWithRetry,
		newWebhookLogStage(),
	)
	return &WebhookPipeline{sequence: seq}
}

// Process runs a fan-out item through the webhook delivery pipeline.
func (p *WebhookPipeline) Process(ctx context.Context, item *FanOutItem) (*FanOutItem, error) {
	return p.sequence.Process(ctx, item)
}
