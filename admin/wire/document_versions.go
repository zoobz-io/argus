// Package wire defines request and response types for the admin API surface.
package wire

import (
	"time"

	"github.com/zoobz-io/argus/models"
)

// AdminDocumentVersionResponse is the admin API response for a document version.
type AdminDocumentVersionResponse struct {
	CreatedAt        time.Time               `json:"created_at" description:"Creation timestamp"`
	ObjectKey        string                  `json:"object_key" description:"MinIO object key"`
	ContentHash      string                  `json:"content_hash" description:"Content hash for dedup"`
	ExtractionStatus models.ExtractionStatus `json:"extraction_status" description:"Extraction status" example:"completed"`
	ID               int64                   `json:"id" description:"Version ID" example:"1"`
	DocumentID       int64                   `json:"document_id" description:"Document ID" example:"1"`
	TenantID         int64                   `json:"tenant_id" description:"Owning tenant ID" example:"1"`
	VersionNumber    int                     `json:"version_number" description:"Version number" example:"1"`
}

// Clone returns a copy of the response.
func (v AdminDocumentVersionResponse) Clone() AdminDocumentVersionResponse {
	return v
}

// AdminDocumentVersionListResponse is the admin API response for listing document versions.
type AdminDocumentVersionListResponse struct {
	Cursor   *int64                         `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Versions []AdminDocumentVersionResponse `json:"versions" description:"List of document versions"`
	Limit    int                            `json:"limit" description:"Page size" example:"20"`
	HasMore  bool                           `json:"has_more" description:"Whether more results exist"`
}

// Clone returns a deep copy of the list response.
func (r AdminDocumentVersionListResponse) Clone() AdminDocumentVersionListResponse {
	c := r
	if r.Versions != nil {
		c.Versions = make([]AdminDocumentVersionResponse, len(r.Versions))
		copy(c.Versions, r.Versions)
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
