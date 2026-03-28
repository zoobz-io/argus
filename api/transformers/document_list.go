package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentsToListResponse converts an offset result of documents to an API list response.
func DocumentsToListResponse(result *models.OffsetResult[models.Document]) wire.DocumentListResponse {
	items := make([]wire.DocumentResponse, len(result.Items))
	for i, d := range result.Items {
		items[i] = DocumentToResponse(d)
	}
	return wire.DocumentListResponse{
		Documents: items,
		Offset:    result.Offset,
		Total:     result.Total,
	}
}
