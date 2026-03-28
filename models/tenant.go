package models

import "time"

// Tenant represents an organizational boundary in the system.
type Tenant struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	Name      string    `json:"name" db:"name" constraints:"notnull"`
	Slug      string    `json:"slug" db:"slug" constraints:"notnull,unique"`
	ID        string    `json:"id" db:"id" constraints:"primarykey"`
}

// GetID returns the tenant's primary key.
func (t Tenant) GetID() string {
	return t.ID
}

// GetCreatedAt returns the tenant's creation timestamp.
func (t Tenant) GetCreatedAt() time.Time {
	return t.CreatedAt
}

// Clone returns a copy of the tenant.
func (t Tenant) Clone() Tenant {
	return t
}
