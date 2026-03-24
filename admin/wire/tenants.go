package wire

import "time"

// AdminTenantResponse is the wire representation of a single tenant.
type AdminTenantResponse struct {
	CreatedAt time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" description:"Last update timestamp"`
	Name      string    `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug      string    `json:"slug" description:"URL-friendly identifier" example:"acme-corp"`
	ID        int64     `json:"id" description:"Tenant ID" example:"1"`
}

// Clone returns a shallow copy of the response.
func (t AdminTenantResponse) Clone() AdminTenantResponse {
	return t
}

// AdminTenantListResponse is the wire representation of a paginated tenant list.
type AdminTenantListResponse struct {
	Cursor  *int64                `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Tenants []AdminTenantResponse `json:"tenants" description:"List of tenants"`
	Limit   int                   `json:"limit" description:"Page size" example:"20"`
	HasMore bool                  `json:"has_more" description:"Whether more results exist"`
}

// Clone returns a deep copy of the response.
func (r AdminTenantListResponse) Clone() AdminTenantListResponse {
	c := r
	if r.Tenants != nil {
		c.Tenants = make([]AdminTenantResponse, len(r.Tenants))
		copy(c.Tenants, r.Tenants)
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
