package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Topics defines admin operations for topics.
type Topics interface {
	// GetTopic retrieves a topic by ID.
	GetTopic(ctx context.Context, id string) (*models.Topic, error)
	// CreateTopic creates a new topic for a tenant.
	CreateTopic(ctx context.Context, tenantID, name, description string) (*models.Topic, error)
	// UpdateTopic updates an existing topic.
	UpdateTopic(ctx context.Context, id, name, description string) (*models.Topic, error)
	// DeleteTopic removes a topic.
	DeleteTopic(ctx context.Context, id string) error
	// ListTopicsByTenant retrieves topics for a specific tenant.
	ListTopicsByTenant(ctx context.Context, tenantID string) ([]*models.Topic, error)
}
