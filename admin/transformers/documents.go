package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// DocumentToAdminResponse converts a document model to an admin API response.
func DocumentToAdminResponse(d *models.Document) wire.AdminDocumentResponse {
	return wire.AdminDocumentResponse{
		ID:               d.ID,
		TenantID:         d.TenantID,
		ProviderID:       d.ProviderID,
		WatchedPathID:    d.WatchedPathID,
		CurrentVersionID: d.CurrentVersionID,
		ExternalID:       d.ExternalID,
		Name:             d.Name,
		MimeType:         d.MimeType,
		CreatedAt:        d.CreatedAt,
		UpdatedAt:        d.UpdatedAt,
	}
}

// DocumentsToAdminList converts a slice of document models to a paginated admin list response.
func DocumentsToAdminList(docs []*models.Document, total, limit, offset int) wire.AdminDocumentListResponse {
	items := make([]wire.AdminDocumentResponse, len(docs))
	for i, d := range docs {
		items[i] = DocumentToAdminResponse(d)
	}
	return wire.AdminDocumentListResponse{
		Documents: items,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}
}
