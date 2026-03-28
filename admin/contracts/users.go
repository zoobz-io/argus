package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Users defines admin operations for users.
type Users interface {
	// GetUser retrieves a user by ID.
	GetUser(ctx context.Context, id string) (*models.User, error)
	// CreateUser creates a new user.
	CreateUser(ctx context.Context, tenantID string, externalID string, email string, displayName string, role models.UserRole) (*models.User, error)
	// UpdateUser updates an existing user.
	UpdateUser(ctx context.Context, id string, email string, displayName string, role models.UserRole, status models.UserStatus) (*models.User, error)
	// DeleteUser removes a user.
	DeleteUser(ctx context.Context, id string) error
	// ListUsers retrieves a paginated list of all users (cross-tenant).
	ListUsers(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.User], error)
}
