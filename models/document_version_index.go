package models

// DocumentVersionIndex is the OpenSearch document type for searchable document content.
type DocumentVersionIndex struct {
	DocumentName string    `json:"document_name"`
	MimeType     string    `json:"mime_type"`
	Content      string    `json:"content"`
	Summary      string    `json:"summary"`
	Embedding    []float32 `json:"embedding"`
	TenantID     int64     `json:"tenant_id"`
	ProviderID   int64     `json:"provider_id"`
	DocumentID   int64     `json:"document_id"`
	VersionID    int64     `json:"version_id"`
}

// Clone returns a deep copy of the index document.
func (d DocumentVersionIndex) Clone() DocumentVersionIndex {
	c := d
	if d.Embedding != nil {
		c.Embedding = make([]float32, len(d.Embedding))
		copy(c.Embedding, d.Embedding)
	}
	return c
}
