package models

import "time"

// WatchedPath represents a path a tenant has registered for monitoring within a provider.
type WatchedPath struct {
	CreatedAt  time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	Path       string    `json:"path" db:"path" constraints:"notnull"`
	ID         int64     `json:"id" db:"id" constraints:"primarykey"`
	TenantID   int64     `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ProviderID int64     `json:"provider_id" db:"provider_id" constraints:"notnull"`
	Active     bool      `json:"active" db:"active" constraints:"notnull" default:"true"`
}

// GetID returns the watched path's primary key.
func (w WatchedPath) GetID() int64 {
	return w.ID
}

// Clone returns a copy of the watched path.
func (w WatchedPath) Clone() WatchedPath {
	return w
}
