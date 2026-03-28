package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionsToListResponse converts an offset result of document versions to an API list response.
func DocumentVersionsToListResponse(result *models.OffsetResult[models.DocumentVersion]) wire.DocumentVersionListResponse {
	items := make([]wire.DocumentVersionResponse, len(result.Items))
	for i, v := range result.Items {
		items[i] = DocumentVersionToResponse(v)
	}
	return wire.DocumentVersionListResponse{
		Versions: items,
		Offset:   result.Offset,
		Total:    result.Total,
	}
}
