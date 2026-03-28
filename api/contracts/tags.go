package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Tags defines the contract for tag operations on the public API surface.
type Tags interface {
	// GetTag retrieves a tag by ID.
	GetTag(ctx context.Context, id string) (*models.Tag, error)
	// CreateTag creates a new tag for a tenant.
	CreateTag(ctx context.Context, tenantID, name, description string) (*models.Tag, error)
	// UpdateTag updates an existing tag.
	UpdateTag(ctx context.Context, id, name, description string) (*models.Tag, error)
	// ListTagsByTenant retrieves all tags for a tenant.
	ListTagsByTenant(ctx context.Context, tenantID string) ([]*models.Tag, error)
}
