package stores

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Tags provides database access for tags.
type Tags struct {
	*sum.Database[models.Tag]
}

// NewTags creates a new tags store.
func NewTags(db *sqlx.DB, renderer astql.Renderer) *Tags {
	return &Tags{
		Database: sum.NewDatabase[models.Tag](db, "tags", renderer),
	}
}

// GetTag retrieves a tag by ID.
func (s *Tags) GetTag(ctx context.Context, id string) (*models.Tag, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateTag creates a new tag.
func (s *Tags) CreateTag(ctx context.Context, tenantID, name, description string) (*models.Tag, error) {
	t := &models.Tag{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
	}
	if err := s.Set(ctx, "", t); err != nil {
		return nil, fmt.Errorf("creating tag: %w", err)
	}
	return t, nil
}

// UpdateTag updates an existing tag.
func (s *Tags) UpdateTag(ctx context.Context, id, name, description string) (*models.Tag, error) {
	t, err := s.GetTag(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	t.Description = description
	if err := s.Set(ctx, id, t); err != nil {
		return nil, fmt.Errorf("updating tag: %w", err)
	}
	return t, nil
}

// DeleteTag removes a tag.
func (s *Tags) DeleteTag(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ListTagsByTenant retrieves all tags for a tenant.
func (s *Tags) ListTagsByTenant(ctx context.Context, tenantID string) ([]*models.Tag, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("name", "ASC").
		Exec(ctx, map[string]any{"tenant_id": tenantID})
}
