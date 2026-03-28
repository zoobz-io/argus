package stores

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// Jobs provides database access for ingestion jobs.
type Jobs struct {
	*sum.Database[models.Job]
}

// NewJobs creates a new jobs store.
func NewJobs(db *sqlx.DB, renderer astql.Renderer) *Jobs {
	return &Jobs{
		Database: sum.NewDatabase[models.Job](db, "jobs", renderer),
	}
}

// CreateJob inserts a new job and returns it.
func (s *Jobs) CreateJob(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error) {
	j := &models.Job{
		ID:         uuid.New().String(),
		VersionID:  versionID,
		DocumentID: documentID,
		TenantID:   tenantID,
		Status:     models.JobPending,
	}
	if err := s.Set(ctx, "", j); err != nil {
		return nil, fmt.Errorf("creating job: %w", err)
	}
	return j, nil
}

// GetJob retrieves a job by ID.
func (s *Jobs) GetJob(ctx context.Context, id string) (*models.Job, error) {
	job, err := s.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("getting job: %w", err)
	}
	return job, nil
}

// UpdateJobStatus sets the status (and optional error) for a job.
func (s *Jobs) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, jobErr *string) error {
	params := map[string]any{"id": id, "status": string(status), "updated_at": time.Now()}
	q := s.Modify().
		Set("status", "status").
		Set("updated_at", "updated_at").
		Where("id", "=", "id")
	if jobErr != nil {
		q = q.Set("error", "error")
		params["error"] = *jobErr
	}
	_, err := q.Exec(ctx, params)
	if err != nil {
		return fmt.Errorf("updating job status: %w", err)
	}
	return nil
}
