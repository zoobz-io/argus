package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentToResponse transforms a Document model to an API response.
func DocumentToResponse(d *models.Document) wire.DocumentResponse {
	return wire.DocumentResponse{
		ID:               d.ID,
		ProviderID:       d.ProviderID,
		CurrentVersionID: d.CurrentVersionID,
		Name:             d.Name,
		MimeType:         d.MimeType,
	}
}

// DocumentsToResponse transforms a slice of Document models to responses.
func DocumentsToResponse(docs []*models.Document) []wire.DocumentResponse {
	result := make([]wire.DocumentResponse, len(docs))
	for i, d := range docs {
		result[i] = DocumentToResponse(d)
	}
	return result
}
