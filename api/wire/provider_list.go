package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// ProviderListResponse is the API response for a paginated list of providers.
type ProviderListResponse struct {
	Cursor    *int64             `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Providers []ProviderResponse `json:"providers" description:"List of providers"`
	Limit     int                `json:"limit" description:"Page size" example:"20"`
	HasMore   bool               `json:"has_more" description:"Whether more results exist"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *ProviderListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[ProviderListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r ProviderListResponse) Clone() ProviderListResponse {
	c := r
	if r.Providers != nil {
		c.Providers = make([]ProviderResponse, len(r.Providers))
		copy(c.Providers, r.Providers)
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
