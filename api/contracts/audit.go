package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// AuditLog defines the contract for audit log operations on the public API surface.
type AuditLog interface {
	// Search queries domain events scoped to a tenant.
	Search(ctx context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error)
}
