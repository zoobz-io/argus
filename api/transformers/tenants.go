package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// TenantToResponse transforms a Tenant model to an API response.
func TenantToResponse(t *models.Tenant) wire.TenantResponse {
	return wire.TenantResponse{
		ID:   t.ID,
		Name: t.Name,
		Slug: t.Slug,
	}
}

// TenantsToResponse transforms a slice of Tenant models to responses.
func TenantsToResponse(tenants []*models.Tenant) []wire.TenantResponse {
	result := make([]wire.TenantResponse, len(tenants))
	for i, t := range tenants {
		result[i] = TenantToResponse(t)
	}
	return result
}
