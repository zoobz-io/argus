package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// IngestEnqueuer defines the contract for async ingestion.
type IngestEnqueuer interface {
	// Enqueue creates a job and publishes to the ingestion stream.
	// tenantID scopes the operation — returns not-found if the version belongs to a different tenant.
	Enqueue(ctx context.Context, versionID, tenantID string) (*models.Job, error)
}

// JobReader defines the contract for reading job status scoped to a tenant.
type JobReader interface {
	// GetJobByTenant retrieves a job by ID scoped to a tenant.
	// Returns not-found if the job does not exist or belongs to a different tenant.
	GetJobByTenant(ctx context.Context, id, tenantID string) (*models.Job, error)
}
