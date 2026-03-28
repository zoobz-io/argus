package models

import "time"

// DocumentVersionIndex is the OpenSearch document type for searchable document content.
type DocumentVersionIndex struct {
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DocumentName string    `json:"document_name"`
	MimeType     string    `json:"mime_type"`
	Content      string    `json:"content"`
	Summary      string    `json:"summary"`
	Language     string    `json:"language"`
	TenantID     string    `json:"tenant_id"`
	ProviderID   string    `json:"provider_id"`
	DocumentID   string    `json:"document_id"`
	VersionID    string    `json:"version_id"`
	Topics       []string  `json:"topics"`
	Tags         []string  `json:"tags"`
	Embedding    []float32 `json:"embedding"`
}

// Clone returns a deep copy of the index document.
func (d DocumentVersionIndex) Clone() DocumentVersionIndex {
	c := d
	if d.Topics != nil {
		c.Topics = make([]string, len(d.Topics))
		copy(c.Topics, d.Topics)
	}
	if d.Tags != nil {
		c.Tags = make([]string, len(d.Tags))
		copy(c.Tags, d.Tags)
	}
	if d.Embedding != nil {
		c.Embedding = make([]float32, len(d.Embedding))
		copy(c.Embedding, d.Embedding)
	}
	return c
}
