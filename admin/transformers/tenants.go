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

// TenantsToAdminList converts a cursor result of tenants to an admin list response.
func TenantsToAdminList(result *models.CursorResult[models.Tenant], limit int) wire.AdminTenantListResponse {
	items := make([]wire.AdminTenantResponse, len(result.Items))
	for i, t := range result.Items {
		items[i] = TenantToAdminResponse(t)
	}
	return wire.AdminTenantListResponse{
		Tenants: items,
		Cursor:  result.Cursor,
		Limit:   limit,
		HasMore: result.HasMore,
	}
}
