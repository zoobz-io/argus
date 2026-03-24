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

// ProvidersToAdminList converts a cursor result of providers to an admin list response.
func ProvidersToAdminList(result *models.CursorResult[models.Provider], limit int) wire.AdminProviderListResponse {
	items := make([]wire.AdminProviderResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = ProviderToAdminResponse(p)
	}
	return wire.AdminProviderListResponse{
		Providers: items,
		Cursor:    result.Cursor,
		Limit:     limit,
		HasMore:   result.HasMore,
	}
}
