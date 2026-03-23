package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// ProviderToAdminResponse converts a provider model to an admin API response.
func ProviderToAdminResponse(p *models.Provider) wire.AdminProviderResponse {
	return wire.AdminProviderResponse{
		ID:        p.ID,
		TenantID:  p.TenantID,
		Type:      p.Type,
		Name:      p.Name,
		Active:    p.Active,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// ProvidersToAdminList converts a slice of provider models to a paginated admin list response.
func ProvidersToAdminList(providers []*models.Provider, total, limit, offset int) wire.AdminProviderListResponse {
	items := make([]wire.AdminProviderResponse, len(providers))
	for i, p := range providers {
		items[i] = ProviderToAdminResponse(p)
	}
	return wire.AdminProviderListResponse{
		Providers: items,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}
}
