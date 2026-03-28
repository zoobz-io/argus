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

// Topics provides database access for topics.
type Topics struct {
	*sum.Database[models.Topic]
}

// NewTopics creates a new topics store.
func NewTopics(db *sqlx.DB, renderer astql.Renderer) *Topics {
	return &Topics{
		Database: sum.NewDatabase[models.Topic](db, "topics", renderer),
	}
}

// GetTopic retrieves a topic by ID.
func (s *Topics) GetTopic(ctx context.Context, id string) (*models.Topic, error) {
	return s.Select().
		Where("id", "=", "id").
		Exec(ctx, map[string]any{"id": id})
}

// CreateTopic creates a new topic.
func (s *Topics) CreateTopic(ctx context.Context, tenantID, name, description string) (*models.Topic, error) {
	t := &models.Topic{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Name:        name,
		Description: description,
	}
	if err := s.Set(ctx, "", t); err != nil {
		return nil, fmt.Errorf("creating topic: %w", err)
	}
	return t, nil
}

// UpdateTopic updates an existing topic.
func (s *Topics) UpdateTopic(ctx context.Context, id, name, description string) (*models.Topic, error) {
	t, err := s.GetTopic(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Name = name
	t.Description = description
	if err := s.Set(ctx, id, t); err != nil {
		return nil, fmt.Errorf("updating topic: %w", err)
	}
	return t, nil
}

// DeleteTopic removes a topic.
func (s *Topics) DeleteTopic(ctx context.Context, id string) error {
	return s.Delete(ctx, id)
}

// ListTopicsByTenant retrieves all topics for a tenant.
func (s *Topics) ListTopicsByTenant(ctx context.Context, tenantID string) ([]*models.Topic, error) {
	return s.Query().
		Where("tenant_id", "=", "tenant_id").
		OrderBy("name", "ASC").
		Exec(ctx, map[string]any{"tenant_id": tenantID})
}
