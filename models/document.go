// Package models defines domain entities representing core business objects.
package models

import "time"

// Document represents a file being tracked by the system.
type Document struct {
	CreatedAt        time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	CurrentVersionID *string   `json:"current_version_id,omitempty" db:"current_version_id"`
	ExternalID       string    `json:"external_id" db:"external_id" constraints:"notnull"`
	Name             string    `json:"name" db:"name" constraints:"notnull"`
	MimeType         string    `json:"mime_type" db:"mime_type" constraints:"notnull"`
	ObjectKey        string    `json:"object_key" db:"object_key" constraints:"notnull"`
	ID               string    `json:"id" db:"id" constraints:"primarykey"`
	TenantID         string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ProviderID       string    `json:"provider_id" db:"provider_id" constraints:"notnull"`
	WatchedPathID    string    `json:"watched_path_id" db:"watched_path_id" constraints:"notnull"`
}

// GetID returns the document's primary key.
func (d Document) GetID() string {
	return d.ID
}

// GetCreatedAt returns the document's creation timestamp.
func (d Document) GetCreatedAt() time.Time {
	return d.CreatedAt
}

// Clone returns a deep copy of the document.
func (d Document) Clone() Document {
	c := d
	if d.CurrentVersionID != nil {
		v := *d.CurrentVersionID
		c.CurrentVersionID = &v
	}
	return c
}
