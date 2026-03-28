package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Tags defines admin operations for tags.
type Tags interface {
	// GetTag retrieves a tag by ID.
	GetTag(ctx context.Context, id string) (*models.Tag, error)
	// CreateTag creates a new tag for a tenant.
	CreateTag(ctx context.Context, tenantID, name, description string) (*models.Tag, error)
	// UpdateTag updates an existing tag.
	UpdateTag(ctx context.Context, id, name, description string) (*models.Tag, error)
	// DeleteTag removes a tag.
	DeleteTag(ctx context.Context, id string) error
	// ListTagsByTenant retrieves tags for a specific tenant.
	ListTagsByTenant(ctx context.Context, tenantID string) ([]*models.Tag, error)
}
