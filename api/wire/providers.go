package wire

import (
	"context"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// ProviderResponse is the public API response for a provider.
type ProviderResponse struct {
	Type   models.ProviderType `json:"type" description:"Provider type" example:"google_drive"`
	Name   string              `json:"name" description:"Provider name" example:"My Google Drive"`
	ID     int64               `json:"id" description:"Provider ID" example:"1"`
	Active bool                `json:"active" description:"Whether the provider is active"`
}

// OnSend applies boundary masking before the response is marshaled.
func (p *ProviderResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[ProviderResponse]](ctx)
	masked, err := b.Send(ctx, *p)
	if err != nil {
		return err
	}
	*p = masked
	return nil
}

// Clone returns a copy of the response.
func (p ProviderResponse) Clone() ProviderResponse {
	return p
}

// ProviderCreateRequest is the request body for creating a provider.
type ProviderCreateRequest struct {
	Type        models.ProviderType `json:"type" description:"Provider type" example:"google_drive"`
	Name        string              `json:"name" description:"Human-readable name" example:"My Google Drive"`
	Credentials string              `json:"credentials" description:"Provider credentials (JSON)"`
}

// Validate validates the request fields.
func (r *ProviderCreateRequest) Validate() error {
	return check.All(
		check.Str(string(r.Type), "type").Required().V(),
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
		check.Str(r.Credentials, "credentials").Required().V(),
	).Err()
}

// Clone returns a copy of the request.
func (r ProviderCreateRequest) Clone() ProviderCreateRequest {
	return r
}
