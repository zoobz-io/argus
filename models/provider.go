package models

import (
	"context"
	"time"

	"github.com/zoobz-io/sum"
)

// ProviderType identifies a cloud storage provider.
type ProviderType string

// Supported provider types.
const (
	ProviderGoogleDrive ProviderType = "google_drive"
	ProviderOneDrive    ProviderType = "onedrive"
	ProviderDropbox     ProviderType = "dropbox"
	ProviderS3          ProviderType = "s3"
	ProviderGCS         ProviderType = "gcs"
	ProviderAzureBlob   ProviderType = "azure_blob"
)

// Provider represents a tenant's connection to a cloud storage service.
type Provider struct {
	CreatedAt   time.Time    `json:"created_at" db:"created_at" default:"now()"`
	UpdatedAt   time.Time    `json:"updated_at" db:"updated_at" default:"now()"`
	Type        ProviderType `json:"type" db:"type" constraints:"notnull"`
	Name        string       `json:"name" db:"name" constraints:"notnull"`
	Credentials string       `json:"-" db:"credentials" constraints:"notnull" store.encrypt:"aes" load.decrypt:"aes"`
	ID          int64        `json:"id" db:"id" constraints:"primarykey"`
	TenantID    int64        `json:"tenant_id" db:"tenant_id" constraints:"notnull"`
	Active      bool         `json:"active" db:"active" constraints:"notnull" default:"true"`
}

// BeforeSave encrypts sensitive fields before persistence.
func (p *Provider) BeforeSave(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[Provider]](ctx)
	stored, err := b.Store(ctx, *p)
	if err != nil {
		return err
	}
	*p = stored
	return nil
}

// AfterLoad decrypts sensitive fields after loading.
func (p *Provider) AfterLoad(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[Provider]](ctx)
	loaded, err := b.Load(ctx, *p)
	if err != nil {
		return err
	}
	*p = loaded
	return nil
}

// GetID returns the provider's primary key.
func (p Provider) GetID() int64 {
	return p.ID
}

// Clone returns a deep copy of the provider.
func (p Provider) Clone() Provider {
	return p
}
