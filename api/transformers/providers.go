package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// ProviderToResponse transforms a Provider model to an API response.
func ProviderToResponse(p *models.Provider) wire.ProviderResponse {
	return wire.ProviderResponse{
		ID:     p.ID,
		Type:   p.Type,
		Name:   p.Name,
		Active: p.Active,
	}
}

// ProvidersToResponse transforms a slice of Provider models to responses.
func ProvidersToResponse(providers []*models.Provider) []wire.ProviderResponse {
	result := make([]wire.ProviderResponse, len(providers))
	for i, p := range providers {
		result[i] = ProviderToResponse(p)
	}
	return result
}

// ApplyProviderCreate applies a create request to a Provider model.
func ApplyProviderCreate(req wire.ProviderCreateRequest, p *models.Provider) {
	p.Type = req.Type
	p.Name = req.Name
	p.Credentials = req.Credentials
}
