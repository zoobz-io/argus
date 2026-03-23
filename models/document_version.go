package models

import "time"

// ExtractionStatus represents the state of content extraction for a document version.
type ExtractionStatus string

// Extraction status values.
const (
	ExtractionPending    ExtractionStatus = "pending"
	ExtractionProcessing ExtractionStatus = "processing"
	ExtractionCompleted  ExtractionStatus = "completed"
	ExtractionFailed     ExtractionStatus = "failed"
)

// DocumentVersion represents a specific revision of a document.
type DocumentVersion struct {
	CreatedAt        time.Time        `json:"created_at" db:"created_at" default:"now()"`
	ObjectKey        string           `json:"object_key" db:"object_key" constraints:"notnull"`
	ContentHash      string           `json:"content_hash" db:"content_hash" constraints:"notnull"`
	ExtractionStatus ExtractionStatus `json:"extraction_status" db:"extraction_status" constraints:"notnull" default:"'pending'"`
	ID               int64            `json:"id" db:"id" constraints:"primarykey"`
	DocumentID       int64            `json:"document_id" db:"document_id" constraints:"notnull"`
	TenantID         int64            `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	VersionNumber    int              `json:"version_number" db:"version_number" constraints:"notnull"`
}

// Clone returns a deep copy of the document version.
func (v DocumentVersion) Clone() DocumentVersion {
	return v
}
