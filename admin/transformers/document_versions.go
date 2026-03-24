package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionToAdminResponse converts a document version model to an admin API response.
func DocumentVersionToAdminResponse(v *models.DocumentVersion) wire.AdminDocumentVersionResponse {
	return wire.AdminDocumentVersionResponse{
		ID:               v.ID,
		DocumentID:       v.DocumentID,
		TenantID:         v.TenantID,
		VersionNumber:    v.VersionNumber,
		ObjectKey:        v.ObjectKey,
		ContentHash:      v.ContentHash,
		ExtractionStatus: v.ExtractionStatus,
		CreatedAt:        v.CreatedAt,
	}
}

// DocumentVersionsToAdminList converts a cursor result of document versions to an admin list response.
func DocumentVersionsToAdminList(result *models.CursorResult[models.DocumentVersion], limit int) wire.AdminDocumentVersionListResponse {
	items := make([]wire.AdminDocumentVersionResponse, len(result.Items))
	for i, v := range result.Items {
		items[i] = DocumentVersionToAdminResponse(v)
	}
	return wire.AdminDocumentVersionListResponse{
		Versions: items,
		Cursor:   result.Cursor,
		Limit:    limit,
		HasMore:  result.HasMore,
	}
}
