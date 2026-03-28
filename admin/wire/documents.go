package wire

import "time"

// AdminDocumentResponse is the admin API response for a document.
type AdminDocumentResponse struct {
	CreatedAt        time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt        time.Time `json:"updated_at" description:"Last update timestamp"`
	CurrentVersionID *string   `json:"current_version_id,omitempty" description:"Current version ID"`
	ExternalID       string    `json:"external_id" description:"External provider ID"`
	Name             string    `json:"name" description:"Document name" example:"report.pdf"`
	MimeType         string    `json:"mime_type" description:"MIME type" example:"application/pdf"`
	ID               string    `json:"id" description:"Document ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID         string    `json:"tenant_id" description:"Owning tenant ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProviderID       string    `json:"provider_id" description:"Provider ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	WatchedPathID    string    `json:"watched_path_id" description:"Watched path ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Clone returns a deep copy of the response.
func (d AdminDocumentResponse) Clone() AdminDocumentResponse {
	c := d
	if d.CurrentVersionID != nil {
		v := *d.CurrentVersionID
		c.CurrentVersionID = &v
	}
	return c
}

// AdminDocumentListResponse is the admin API response for listing documents.
type AdminDocumentListResponse struct {
	Documents []AdminDocumentResponse `json:"documents" description:"List of documents"`
	Offset    int                     `json:"offset" description:"Number of results skipped"`
	Limit     int                     `json:"limit" description:"Page size" example:"20"`
	Total     int64                   `json:"total" description:"Total number of results"`
}

// Clone returns a deep copy of the list response.
func (r AdminDocumentListResponse) Clone() AdminDocumentListResponse {
	c := r
	if r.Documents != nil {
		c.Documents = make([]AdminDocumentResponse, len(r.Documents))
		for i, d := range r.Documents {
			c.Documents[i] = d.Clone()
		}
	}
	return c
}
