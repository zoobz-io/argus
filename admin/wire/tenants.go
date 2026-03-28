package wire

import "time"

// AdminTenantResponse is the wire representation of a single tenant.
type AdminTenantResponse struct {
	CreatedAt time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt time.Time `json:"updated_at" description:"Last update timestamp"`
	Name      string    `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug      string    `json:"slug" description:"URL-friendly identifier" example:"acme-corp"`
	ID        string    `json:"id" description:"Tenant ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Clone returns a shallow copy of the response.
func (t AdminTenantResponse) Clone() AdminTenantResponse {
	return t
}

// AdminTenantListResponse is the wire representation of a paginated tenant list.
type AdminTenantListResponse struct {
	Tenants []AdminTenantResponse `json:"tenants" description:"List of tenants"`
	Offset  int                   `json:"offset" description:"Number of results skipped"`
	Limit   int                   `json:"limit" description:"Page size" example:"20"`
	Total   int64                 `json:"total" description:"Total number of results"`
}

// Clone returns a deep copy of the response.
func (r AdminTenantListResponse) Clone() AdminTenantListResponse {
	c := r
	if r.Tenants != nil {
		c.Tenants = make([]AdminTenantResponse, len(r.Tenants))
		copy(c.Tenants, r.Tenants)
	}
	return c
}
