package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// ProvidersToListResponse converts a cursor result of providers to an API list response.
func ProvidersToListResponse(result *models.CursorResult[models.Provider], limit int) wire.ProviderListResponse {
	items := make([]wire.ProviderResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = ProviderToResponse(p)
	}
	return wire.ProviderListResponse{
		Providers: items,
		Cursor:    result.Cursor,
		Limit:     limit,
		HasMore:   result.HasMore,
	}
}
