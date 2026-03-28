package ingest

import (
	"context"
	"fmt"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

// Enqueuer creates jobs and emits the queue signal for herald dispatch.
type Enqueuer struct{}

// NewEnqueuer creates a new Enqueuer.
func NewEnqueuer() *Enqueuer {
	return &Enqueuer{}
}

// Enqueue creates a pending job and emits the queue signal.
// tenantID scopes the operation — returns not-found if the version belongs to a different tenant.
func (e *Enqueuer) Enqueue(ctx context.Context, versionID, tenantID string) (*models.Job, error) {
	versions := sum.MustUse[intcontracts.IngestVersions](ctx)
	documents := sum.MustUse[intcontracts.IngestDocuments](ctx)
	jobs := sum.MustUse[intcontracts.IngestJobs](ctx)

	version, err := versions.GetDocumentVersion(ctx, versionID)
	if err != nil {
		return nil, fmt.Errorf("fetching document version: %w", err)
	}

	if version.TenantID != tenantID {
		return nil, fmt.Errorf("fetching document version: not found")
	}

	if _, err := documents.GetDocument(ctx, version.DocumentID); err != nil {
		return nil, fmt.Errorf("fetching document: %w", err)
	}

	job, err := jobs.CreateJob(ctx, version.ID, version.DocumentID, version.TenantID)
	if err != nil {
		return nil, fmt.Errorf("creating job: %w", err)
	}

	capitan.Emit(ctx, events.IngestQueueSignal, events.IngestQueueKey.Field(events.IngestMessage{
		JobID:     job.ID,
		VersionID: version.ID,
	}))

	return job, nil
}
