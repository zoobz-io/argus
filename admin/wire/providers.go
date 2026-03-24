package wire

import (
	"time"

	"github.com/zoobz-io/argus/models"
)

// AdminProviderResponse is the wire representation of a single provider.
type AdminProviderResponse struct {
	CreatedAt time.Time           `json:"created_at" description:"Creation timestamp"`
	UpdatedAt time.Time           `json:"updated_at" description:"Last update timestamp"`
	Type      models.ProviderType `json:"type" description:"Provider type" example:"google_drive"`
	Name      string              `json:"name" description:"Provider name" example:"My Google Drive"`
	ID        int64               `json:"id" description:"Provider ID" example:"1"`
	TenantID  int64               `json:"tenant_id" description:"Owning tenant ID" example:"1"`
	Active    bool                `json:"active" description:"Whether the provider is active"`
}

// Clone returns a shallow copy of the response.
func (p AdminProviderResponse) Clone() AdminProviderResponse {
	return p
}

// AdminProviderListResponse is the wire representation of a paginated provider list.
type AdminProviderListResponse struct {
	Cursor    *int64                  `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Providers []AdminProviderResponse `json:"providers" description:"List of providers"`
	Limit     int                     `json:"limit" description:"Page size" example:"20"`
	HasMore   bool                    `json:"has_more" description:"Whether more results exist"`
}

// Clone returns a deep copy of the response.
func (r AdminProviderListResponse) Clone() AdminProviderListResponse {
	c := r
	if r.Providers != nil {
		c.Providers = make([]AdminProviderResponse, len(r.Providers))
		copy(c.Providers, r.Providers)
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
