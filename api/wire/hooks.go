package wire

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// HookCreateRequest is the request body for creating a webhook endpoint.
type HookCreateRequest struct {
	URL string `json:"url" description:"Webhook endpoint URL" example:"https://example.com/webhook"`
}

// Validate validates the request fields.
func (r *HookCreateRequest) Validate() error {
	if err := check.All(
		check.Str(r.URL, "url").Required().V(),
	).Err(); err != nil {
		return err
	}
	return validateWebhookURL(r.URL)
}

// validateWebhookURL enforces HTTPS scheme and blocks private/loopback destinations.
func validateWebhookURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("webhook URL must use HTTPS scheme")
	}
	host := u.Hostname()
	if ip := net.ParseIP(host); ip != nil {
		if ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
			return fmt.Errorf("webhook URL must not target private or loopback addresses")
		}
	}
	return nil
}

// Clone returns a copy of the request.
func (r HookCreateRequest) Clone() HookCreateRequest {
	return r
}

// HookCreateResponse is the API response returned once when a hook is created.
// It includes the Secret which is only exposed at creation time.
// No OnSend masking — the secret must be visible to the caller on create.
type HookCreateResponse struct {
	CreatedAt time.Time `json:"created_at" description:"Creation timestamp"`
	ID        string    `json:"id" description:"Hook ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL       string    `json:"url" description:"Webhook endpoint URL"`
	Secret    string    `json:"secret" description:"Signing secret — shown only on create"`
	Active    bool      `json:"active" description:"Whether the hook is active"`
}

// Clone returns a copy of the response.
func (h HookCreateResponse) Clone() HookCreateResponse {
	return h
}

// HookResponse is the public API response for a webhook endpoint.
// It deliberately omits the Secret field.
type HookResponse struct {
	CreatedAt time.Time `json:"created_at" description:"Creation timestamp"`
	ID        string    `json:"id" description:"Hook ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	URL       string    `json:"url" description:"Webhook endpoint URL" send.redact:"***"`
	Active    bool      `json:"active" description:"Whether the hook is active"`
}

// OnSend applies boundary masking before the response is marshaled.
func (h *HookResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[HookResponse]](ctx)
	masked, err := b.Send(ctx, *h)
	if err != nil {
		return err
	}
	*h = masked
	return nil
}

// Clone returns a copy of the response.
func (h HookResponse) Clone() HookResponse {
	return h
}

// HookListResponse is the public API response for a paginated hook list.
type HookListResponse struct {
	Hooks  []HookResponse `json:"hooks" description:"List of hooks"`
	Offset int            `json:"offset" description:"Number of results skipped"`
	Limit  int            `json:"limit" description:"Page size" example:"20"`
	Total  int64          `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *HookListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[HookListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r HookListResponse) Clone() HookListResponse {
	c := r
	if r.Hooks != nil {
		c.Hooks = make([]HookResponse, len(r.Hooks))
		copy(c.Hooks, r.Hooks)
	}
	return c
}

// DeliveryResponse is the public API response for a webhook delivery attempt.
type DeliveryResponse struct {
	CreatedAt  time.Time `json:"created_at" description:"Delivery timestamp"`
	ID         string    `json:"id" description:"Delivery ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	HookID     string    `json:"hook_id" description:"Associated hook ID"`
	EventID    string    `json:"event_id" description:"Event that triggered the delivery"`
	Error      string    `json:"error,omitempty" description:"Error message if delivery failed"`
	StatusCode int       `json:"status_code" description:"HTTP status code from endpoint"`
	Attempt    int       `json:"attempt" description:"Attempt number"`
}

// OnSend applies boundary masking before the response is marshaled.
func (d *DeliveryResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[DeliveryResponse]](ctx)
	masked, err := b.Send(ctx, *d)
	if err != nil {
		return err
	}
	*d = masked
	return nil
}

// Clone returns a copy of the response.
func (d DeliveryResponse) Clone() DeliveryResponse {
	return d
}

// DeliveryListResponse is the public API response for a paginated delivery list.
type DeliveryListResponse struct {
	Deliveries []DeliveryResponse `json:"deliveries" description:"List of deliveries"`
	Offset     int                `json:"offset" description:"Number of results skipped"`
	Limit      int                `json:"limit" description:"Page size" example:"20"`
	Total      int64              `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *DeliveryListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[DeliveryListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r DeliveryListResponse) Clone() DeliveryListResponse {
	c := r
	if r.Deliveries != nil {
		c.Deliveries = make([]DeliveryResponse, len(r.Deliveries))
		copy(c.Deliveries, r.Deliveries)
	}
	return c
}
