package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// AuditLog defines the contract for audit log operations on the admin API surface.
type AuditLog interface {
	// Search queries domain events across tenants.
	Search(ctx context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error)
}
