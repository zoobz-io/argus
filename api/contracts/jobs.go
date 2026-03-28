package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// IngestEnqueuer defines the contract for async ingestion.
type IngestEnqueuer interface {
	// Enqueue creates a job and publishes to the ingestion stream.
	Enqueue(ctx context.Context, versionID string) (*models.Job, error)
}

// JobReader defines the contract for reading job status.
type JobReader interface {
	// GetJob retrieves a job by ID.
	GetJob(ctx context.Context, id string) (*models.Job, error)
}
