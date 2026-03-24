package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionsToListResponse converts a cursor result of document versions to an API list response.
func DocumentVersionsToListResponse(result *models.CursorResult[models.DocumentVersion], limit int) wire.DocumentVersionListResponse {
	items := make([]wire.DocumentVersionResponse, len(result.Items))
	for i, v := range result.Items {
		items[i] = DocumentVersionToResponse(v)
	}
	return wire.DocumentVersionListResponse{
		Versions: items,
		Cursor:   result.Cursor,
		Limit:    limit,
		HasMore:  result.HasMore,
	}
}
