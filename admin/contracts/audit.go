package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// AuditLog defines the contract for audit log operations on the admin API surface.
type AuditLog interface {
	// SearchAll queries domain events across all tenants.
	SearchAll(ctx context.Context, params models.DomainEventSearchParams) (*models.OffsetResult[models.DomainEvent], error)
}
