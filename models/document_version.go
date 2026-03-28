package models

import "time"

// DocumentVersion represents a specific revision of a document.
type DocumentVersion struct {
	CreatedAt     time.Time `json:"created_at" db:"created_at" default:"now()"`
	ContentHash   string    `json:"content_hash" db:"content_hash" constraints:"notnull"`
	ID            string    `json:"id" db:"id" constraints:"primarykey"`
	DocumentID    string    `json:"document_id" db:"document_id" constraints:"notnull"`
	TenantID      string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	VersionNumber int       `json:"version_number" db:"version_number" constraints:"notnull"`
}

// GetID returns the document version's primary key.
func (v DocumentVersion) GetID() string {
	return v.ID
}

// GetCreatedAt returns the document version's creation timestamp.
func (v DocumentVersion) GetCreatedAt() time.Time {
	return v.CreatedAt
}

// Clone returns a deep copy of the document version.
func (v DocumentVersion) Clone() DocumentVersion {
	return v
}
