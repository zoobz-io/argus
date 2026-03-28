package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// ProviderListResponse is the API response for a paginated list of providers.
type ProviderListResponse struct {
	Providers []ProviderResponse `json:"providers" description:"List of providers"`
	Offset    int                `json:"offset" description:"Number of results skipped"`
	Limit     int                `json:"limit" description:"Page size" example:"20"`
	Total     int64              `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *ProviderListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[ProviderListResponse]](ctx)
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
	return c
}
