package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentsToListResponse converts a cursor result of documents to an API list response.
func DocumentsToListResponse(result *models.CursorResult[models.Document], limit int) wire.DocumentListResponse {
	items := make([]wire.DocumentResponse, len(result.Items))
	for i, d := range result.Items {
		items[i] = DocumentToResponse(d)
	}
	return wire.DocumentListResponse{
		Documents: items,
		Cursor:    result.Cursor,
		Limit:     limit,
		HasMore:   result.HasMore,
	}
}
