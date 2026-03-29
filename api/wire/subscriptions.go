package wire

import (
	"context"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// SubscriptionRequest is the request body for creating a subscription.
type SubscriptionRequest struct {
	EventType         string                     `json:"event_type" description:"Event type to subscribe to" example:"ingest.completed"`
	Channel           models.SubscriptionChannel `json:"channel" description:"Delivery channel" example:"inbox"`
	WebhookEndpointID string                     `json:"webhook_endpoint_id,omitempty" description:"Optional webhook endpoint ID for webhook channel"`
}

// Validate validates the request fields.
func (r *SubscriptionRequest) Validate() error {
	return check.All(
		check.Str(r.EventType, "event_type").Required().V(),
		check.Str(string(r.Channel), "channel").Required().V(),
	).Err()
}

// Clone returns a copy of the request.
func (r SubscriptionRequest) Clone() SubscriptionRequest {
	return r
}

// SubscriptionResponse is the public API response for a subscription.
type SubscriptionResponse struct {
	Channel   models.SubscriptionChannel `json:"channel" description:"Delivery channel" example:"inbox"`
	ID        string                     `json:"id" description:"Subscription ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID    string                     `json:"user_id" description:"Subscriber user ID"`
	EventType string                     `json:"event_type" description:"Event type" example:"ingest.completed"`
}

// OnSend applies boundary masking before the response is marshaled.
func (s *SubscriptionResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[SubscriptionResponse]](ctx)
	masked, err := b.Send(ctx, *s)
	if err != nil {
		return err
	}
	*s = masked
	return nil
}

// Clone returns a copy of the response.
func (s SubscriptionResponse) Clone() SubscriptionResponse {
	return s
}

// SubscriptionListResponse is the public API response for a paginated subscription list.
type SubscriptionListResponse struct {
	Subscriptions []SubscriptionResponse `json:"subscriptions" description:"List of subscriptions"`
	Offset        int                    `json:"offset" description:"Number of results skipped"`
	Limit         int                    `json:"limit" description:"Page size" example:"20"`
	Total         int64                  `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *SubscriptionListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[SubscriptionListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r SubscriptionListResponse) Clone() SubscriptionListResponse {
	c := r
	if r.Subscriptions != nil {
		c.Subscriptions = make([]SubscriptionResponse, len(r.Subscriptions))
		copy(c.Subscriptions, r.Subscriptions)
	}
	return c
}
