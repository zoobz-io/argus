package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DomainEventIndexer defines indexing operations for the unified domain event system.
type DomainEventIndexer interface {
	Index(ctx context.Context, event *models.DomainEvent) error
}
