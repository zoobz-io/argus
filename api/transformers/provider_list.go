package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// ProvidersToListResponse converts an offset result of providers to an API list response.
func ProvidersToListResponse(result *models.OffsetResult[models.Provider]) wire.ProviderListResponse {
	items := make([]wire.ProviderResponse, len(result.Items))
	for i, p := range result.Items {
		items[i] = ProviderToResponse(p)
	}
	return wire.ProviderListResponse{
		Providers: items,
		Offset:    result.Offset,
		Total:     result.Total,
	}
}
