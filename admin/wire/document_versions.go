// Package wire defines request and response types for the admin API surface.
package wire

import "time"

// AdminDocumentVersionResponse is the admin API response for a document version.
type AdminDocumentVersionResponse struct {
	CreatedAt     time.Time `json:"created_at" description:"Creation timestamp"`
	ContentHash   string    `json:"content_hash" description:"Content hash for dedup"`
	ID            string    `json:"id" description:"Version ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	DocumentID    string    `json:"document_id" description:"Document ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID      string    `json:"tenant_id" description:"Owning tenant ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	VersionNumber int       `json:"version_number" description:"Version number" example:"1"`
}

// Clone returns a copy of the response.
func (v AdminDocumentVersionResponse) Clone() AdminDocumentVersionResponse {
	return v
}

// AdminDocumentVersionListResponse is the admin API response for listing document versions.
type AdminDocumentVersionListResponse struct {
	Versions []AdminDocumentVersionResponse `json:"versions" description:"List of document versions"`
	Offset   int                            `json:"offset" description:"Number of results skipped"`
	Limit    int                            `json:"limit" description:"Page size" example:"20"`
	Total    int64                          `json:"total" description:"Total number of results"`
}

// Clone returns a deep copy of the list response.
func (r AdminDocumentVersionListResponse) Clone() AdminDocumentVersionListResponse {
	c := r
	if r.Versions != nil {
		c.Versions = make([]AdminDocumentVersionResponse, len(r.Versions))
		copy(c.Versions, r.Versions)
	}
	return c
}
