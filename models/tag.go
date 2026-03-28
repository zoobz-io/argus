package models

import "time"

// Tag represents a granular label defined by a tenant.
// Tags are used as bounded vocabulary for LLM extraction during ingestion.
type Tag struct {
	CreatedAt   time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	Name        string    `json:"name" db:"name" constraints:"notnull"`
	Description string    `json:"description" db:"description" constraints:"notnull"`
	ID          string    `json:"id" db:"id" constraints:"primarykey"`
	TenantID    string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
}

// GetID returns the tag's primary key.
func (t Tag) GetID() string {
	return t.ID
}

// GetCreatedAt returns the tag's creation timestamp.
func (t Tag) GetCreatedAt() time.Time {
	return t.CreatedAt
}

// Clone returns a copy of the tag.
func (t Tag) Clone() Tag {
	return t
}
