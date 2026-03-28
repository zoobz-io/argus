package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Users defines the contract for user operations on the public API surface.
type Users interface {
	// GetUserByExternalID retrieves a user by their external identity provider ID.
	GetUserByExternalID(ctx context.Context, externalID string) (*models.User, error)
	// ListUsersByTenant retrieves a paginated list of users for a tenant.
	ListUsersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.User], error)
}
