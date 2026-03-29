package models

import "time"

// Delivery records the result of a single webhook delivery attempt.
type Delivery struct {
	CreatedAt  time.Time `json:"created_at" db:"created_at" default:"now()"`
	Error      *string   `json:"error,omitempty" db:"error"`
	ID         string    `json:"id" db:"id" constraints:"primarykey"`
	HookID     string    `json:"hook_id" db:"hook_id" constraints:"notnull"`
	EventID    string    `json:"event_id" db:"event_id" constraints:"notnull"`
	TenantID   string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	StatusCode int       `json:"status_code" db:"status_code" constraints:"notnull"`
	Attempt    int       `json:"attempt" db:"attempt" constraints:"notnull"`
}

// GetID returns the delivery's primary key.
func (d Delivery) GetID() string {
	return d.ID
}

// GetCreatedAt returns the delivery's creation timestamp.
func (d Delivery) GetCreatedAt() time.Time {
	return d.CreatedAt
}

// Clone returns a deep copy of the delivery.
func (d Delivery) Clone() Delivery {
	c := d
	if d.Error != nil {
		e := *d.Error
		c.Error = &e
	}
	return c
}
