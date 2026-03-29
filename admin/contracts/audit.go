package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// AuditLog defines the contract for audit log operations on the admin API surface.
type AuditLog interface {
	// Search queries audit entries across tenants.
	Search(ctx context.Context, params models.AuditSearchParams) (*models.OffsetResult[models.AuditEntry], error)
}
