package wire

import "github.com/zoobz-io/check"

// AdminTenantCreateRequest is the request body for creating a tenant.
type AdminTenantCreateRequest struct {
	Name string `json:"name" description:"Tenant name" example:"Acme Corp"`
	Slug string `json:"slug" description:"URL-friendly identifier" example:"acme-corp"`
}

// Validate checks that all required fields are present.
func (r *AdminTenantCreateRequest) Validate() error {
	return check.All(
		check.Str(r.Name, "name").Required().MaxLen(255).V(),
		check.Str(r.Slug, "slug").Required().MaxLen(255).V(),
	).Err()
}

// Clone returns a copy of the request.
func (r AdminTenantCreateRequest) Clone() AdminTenantCreateRequest {
	return r
}
