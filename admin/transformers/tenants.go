// Package transformers provides functions to convert domain models into admin API wire types.
package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// TenantToAdminResponse converts a tenant model to an admin API response.
func TenantToAdminResponse(t *models.Tenant) wire.AdminTenantResponse {
	return wire.AdminTenantResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}
}

// TenantsToAdminList converts a slice of tenant models to a paginated admin list response.
func TenantsToAdminList(tenants []*models.Tenant, total, limit, offset int) wire.AdminTenantListResponse {
	items := make([]wire.AdminTenantResponse, len(tenants))
	for i, t := range tenants {
		items[i] = TenantToAdminResponse(t)
	}
	return wire.AdminTenantListResponse{
		Tenants: items,
		Total:   total,
		Limit:   limit,
		Offset:  offset,
	}
}
