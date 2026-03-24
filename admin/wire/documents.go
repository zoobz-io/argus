package wire

import "time"

// AdminDocumentResponse is the admin API response for a document.
type AdminDocumentResponse struct {
	CreatedAt        time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt        time.Time `json:"updated_at" description:"Last update timestamp"`
	CurrentVersionID *int64    `json:"current_version_id,omitempty" description:"Current version ID"`
	ExternalID       string    `json:"external_id" description:"External provider ID"`
	Name             string    `json:"name" description:"Document name" example:"report.pdf"`
	MimeType         string    `json:"mime_type" description:"MIME type" example:"application/pdf"`
	ID               int64     `json:"id" description:"Document ID" example:"1"`
	TenantID         int64     `json:"tenant_id" description:"Owning tenant ID" example:"1"`
	ProviderID       int64     `json:"provider_id" description:"Provider ID" example:"1"`
	WatchedPathID    int64     `json:"watched_path_id" description:"Watched path ID" example:"1"`
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
	Cursor    *int64                  `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Documents []AdminDocumentResponse `json:"documents" description:"List of documents"`
	Limit     int                     `json:"limit" description:"Page size" example:"20"`
	HasMore   bool                    `json:"has_more" description:"Whether more results exist"`
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
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
