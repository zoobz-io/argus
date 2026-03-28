package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// TenantResponse is the public API response for a tenant.
type TenantResponse struct {
	Name string `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug string `json:"slug" description:"URL-friendly identifier" example:"acme-corp"`
	ID   string `json:"id" description:"Tenant ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OnSend applies boundary masking before the response is marshaled.
func (t *TenantResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[TenantResponse]](ctx)
	masked, err := b.Send(ctx, *t)
	if err != nil {
		return err
	}
	*t = masked
	return nil
}

// Clone returns a copy of the response.
func (t TenantResponse) Clone() TenantResponse {
	return t
}
