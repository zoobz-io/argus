package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionToAdminResponse converts a document version model to an admin API response.
func DocumentVersionToAdminResponse(v *models.DocumentVersion) wire.AdminDocumentVersionResponse {
	return wire.AdminDocumentVersionResponse{
		ID:            v.ID,
		DocumentID:    v.DocumentID,
		TenantID:      v.TenantID,
		VersionNumber: v.VersionNumber,
		ContentHash:   v.ContentHash,
		CreatedAt:     v.CreatedAt,
	}
}

// DocumentVersionsToAdminList converts an offset result of document versions to an admin list response.
func DocumentVersionsToAdminList(result *models.OffsetResult[models.DocumentVersion]) wire.AdminDocumentVersionListResponse {
	items := make([]wire.AdminDocumentVersionResponse, len(result.Items))
	for i, v := range result.Items {
		items[i] = DocumentVersionToAdminResponse(v)
	}
	return wire.AdminDocumentVersionListResponse{
		Versions: items,
		Offset:   result.Offset,
		Total:    result.Total,
	}
}
