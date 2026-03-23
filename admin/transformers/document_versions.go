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

// DocumentVersionsToAdminList converts a slice of document version models to a paginated admin list response.
func DocumentVersionsToAdminList(versions []*models.DocumentVersion, total, limit, offset int) wire.AdminDocumentVersionListResponse {
	items := make([]wire.AdminDocumentVersionResponse, len(versions))
	for i, v := range versions {
		items[i] = DocumentVersionToAdminResponse(v)
	}
	return wire.AdminDocumentVersionListResponse{
		Versions: items,
		Total:    total,
		Limit:    limit,
		Offset:   offset,
	}
}
