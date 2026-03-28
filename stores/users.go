package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Users provides database access for users.
type Users struct {
	*sum.Database[models.User]
}

// NewUsers creates a new users store.
func NewUsers(db *sqlx.DB, renderer astql.Renderer) *Users {
	return &Users{
		Database: sum.NewDatabase[models.User](db, "users", renderer),
	}
}

// GetUser retrieves a user by ID.
func (s *Users) GetUser(ctx context.Context, id string) (*models.User, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// GetUserByExternalID retrieves a user by external ID.
func (s *Users) GetUserByExternalID(ctx context.Context, externalID string) (*models.User, error) {
	return s.Select().
		Where("external_id", "=", "external_id").
		Exec(ctx, map[string]any{"external_id": externalID})
}

// GetUserByTenant retrieves a user by ID scoped to a tenant.
// Returns not-found if the user does not exist or belongs to a different tenant.
func (s *Users) GetUserByTenant(ctx context.Context, id, tenantID string) (*models.User, error) {
	params := map[string]any{"id": id, "tenant_id": tenantID}
	q := s.Query().
		Where("id", "=", "id").
		Where("tenant_id", "=", "tenant_id").
		Limit(1)
	results, err := q.Exec(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("getting user by tenant: %w", err)
	}
	if len(results) == 0 {
		return nil, fmt.Errorf("user not found")
	}
	return results[0], nil
}

// ListUsersByTenant retrieves users for a specific tenant using offset/limit pagination.
func (s *Users) ListUsersByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.User], error) {
	items, err := s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, map[string]any{"tenant_id": tenantID})
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.User]{Items: items, Offset: page.Offset}, nil
}

// ListUsers retrieves a paginated list of all users.
func (s *Users) ListUsers(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.User], error) {
	items, err := s.Query().
		OrderBy("created_at", "ASC").
		OrderBy("id", "ASC").
		Limit(page.PageSize()).
		Offset(page.Offset).
		Exec(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &models.OffsetResult[models.User]{Items: items, Offset: page.Offset}, nil
}

// CreateUser creates a new user.
func (s *Users) CreateUser(ctx context.Context, tenantID, externalID, email, displayName string, role models.UserRole) (*models.User, error) {
	u := &models.User{
		ID:          uuid.New().String(),
		ExternalID:  externalID,
		TenantID:    tenantID,
		Email:       email,
		DisplayName: displayName,
		Role:        role,
		Status:      models.UserStatusActive,
	}
	if err := s.Set(ctx, "", u); err != nil {
		return nil, fmt.Errorf("creating user: %w", err)
	}
	return u, nil
}

// UpdateUser updates an existing user.
func (s *Users) UpdateUser(ctx context.Context, id, email, displayName string, role models.UserRole, status models.UserStatus) (*models.User, error) {
	u, err := s.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	u.Email = email
	u.DisplayName = displayName
	u.Role = role
	u.Status = status
	if err := s.Set(ctx, id, u); err != nil {
		return nil, fmt.Errorf("updating user: %w", err)
	}
	return u, nil
}

// DeleteUser removes a user.
func (s *Users) DeleteUser(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// UpsertFromClaims inserts a user or updates last_seen_at on conflict with external_id.
func (s *Users) UpsertFromClaims(ctx context.Context, externalID, tenantID, email, displayName string) (*models.User, error) {
	now := time.Now()
	u := &models.User{
		ID:          uuid.New().String(),
		ExternalID:  externalID,
		TenantID:    tenantID,
		Email:       email,
		DisplayName: displayName,
		Role:        models.UserRoleViewer,
		Status:      models.UserStatusActive,
		LastSeenAt:  &now,
	}
	result, err := s.Insert().
		OnConflict("external_id").
		DoUpdate().
		Set("last_seen_at", "last_seen_at").
		Set("email", "email").
		Set("display_name", "display_name").
		Set("updated_at", "updated_at").
		Exec(ctx, u)
	if err != nil {
		return nil, fmt.Errorf("upserting user from claims: %w", err)
	}
	return result, nil
}
