package models

import "time"

// Topic represents a broad thematic category defined by a tenant.
// Topics are used as bounded vocabulary for LLM extraction during ingestion.
type Topic struct {
	CreatedAt   time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	Name        string    `json:"name" db:"name" constraints:"notnull"`
	Description string    `json:"description" db:"description" constraints:"notnull"`
	ID          string    `json:"id" db:"id" constraints:"primarykey"`
	TenantID    string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
}

// GetID returns the topic's primary key.
func (t Topic) GetID() string {
	return t.ID
}

// GetCreatedAt returns the topic's creation timestamp.
func (t Topic) GetCreatedAt() time.Time {
	return t.CreatedAt
}

// Clone returns a copy of the topic.
func (t Topic) Clone() Topic {
	return t
}
