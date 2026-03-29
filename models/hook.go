package models

import (
	"context"
	"time"

	"github.com/zoobz-io/sum"
)

// Hook represents a webhook endpoint registered by a tenant.
type Hook struct {
	CreatedAt time.Time `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at" default:"now()"`
	URL       string    `json:"url" db:"url" constraints:"notnull"`
	Secret    string    `json:"-" db:"secret" constraints:"notnull" store.encrypt:"aes" load.decrypt:"aes"`
	ID        string    `json:"id" db:"id" constraints:"primarykey"`
	TenantID  string    `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	UserID    string    `json:"user_id" db:"user_id" constraints:"notnull"`
	Active    bool      `json:"active" db:"active" constraints:"notnull" default:"true"`
}

// BeforeSave encrypts sensitive fields before persistence.
func (h *Hook) BeforeSave(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[Hook]](ctx)
	stored, err := b.Store(ctx, *h)
	if err != nil {
		return err
	}
	*h = stored
	return nil
}

// AfterLoad decrypts sensitive fields after loading.
func (h *Hook) AfterLoad(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[Hook]](ctx)
	loaded, err := b.Load(ctx, *h)
	if err != nil {
		return err
	}
	*h = loaded
	return nil
}

// GetID returns the hook's primary key.
func (h Hook) GetID() string {
	return h.ID
}

// GetCreatedAt returns the hook's creation timestamp.
func (h Hook) GetCreatedAt() time.Time {
	return h.CreatedAt
}

// Clone returns a deep copy of the hook.
func (h Hook) Clone() Hook {
	return h
}
