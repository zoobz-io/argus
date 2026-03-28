package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Topics defines the contract for topic operations on the public API surface.
type Topics interface {
	// GetTopic retrieves a topic by ID.
	GetTopic(ctx context.Context, id string) (*models.Topic, error)
	// CreateTopic creates a new topic for a tenant.
	CreateTopic(ctx context.Context, tenantID, name, description string) (*models.Topic, error)
	// UpdateTopic updates an existing topic.
	UpdateTopic(ctx context.Context, id, name, description string) (*models.Topic, error)
	// ListTopicsByTenant retrieves all topics for a tenant.
	ListTopicsByTenant(ctx context.Context, tenantID string) ([]*models.Topic, error)
}
