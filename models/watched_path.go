package models

import "time"

// WatchedPath represents a path a tenant has registered for monitoring within a provider.
type WatchedPath struct {
	CreatedAt  time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	SyncState  *string   `json:"-" db:"sync_state" store.encrypt:"aes" load.decrypt:"aes"`
	Path       string    `json:"path" db:"path" constraints:"notnull"`
	ID         string    `json:"id" db:"id" constraints:"primarykey"`
	TenantID   string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	ProviderID string    `json:"provider_id" db:"provider_id" constraints:"notnull"`
	Active     bool      `json:"active" db:"active" constraints:"notnull" default:"true"`
}

// GetID returns the watched path's primary key.
func (w WatchedPath) GetID() string {
	return w.ID
}

// GetCreatedAt returns the watched path's creation timestamp.
func (w WatchedPath) GetCreatedAt() time.Time {
	return w.CreatedAt
}

// Clone returns a deep copy of the watched path.
func (w WatchedPath) Clone() WatchedPath {
	c := w
	if w.SyncState != nil {
		s := *w.SyncState
		c.SyncState = &s
	}
	return c
}
