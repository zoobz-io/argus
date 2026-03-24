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

// DocumentsToAdminList converts a cursor result of documents to an admin list response.
func DocumentsToAdminList(result *models.CursorResult[models.Document], limit int) wire.AdminDocumentListResponse {
	items := make([]wire.AdminDocumentResponse, len(result.Items))
	for i, d := range result.Items {
		items[i] = DocumentToAdminResponse(d)
	}
	return wire.AdminDocumentListResponse{
		Documents: items,
		Cursor:    result.Cursor,
		Limit:     limit,
		HasMore:   result.HasMore,
	}
}
