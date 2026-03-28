//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"

	soytesting "github.com/zoobz-io/soy/testing"
)

func newTestJobs(t *testing.T, mock *soytesting.MockDB) *Jobs {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewJobs(mock.DB(), astqlpg.New())
}

func TestJobs_CreateJob(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Set uses INSERT...RETURNING, which is a query that returns rows.
	mock.ExpectQuery().WithRows([]models.Job{
		{ID: "job-1", Status: models.JobPending, VersionID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", CreatedAt: ts, UpdatedAt: ts},
	})

	job, err := store.CreateJob(context.Background(), "ver-1", "doc-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.Status != models.JobPending {
		t.Errorf("Status: got %q, want %q", job.Status, models.JobPending)
	}
	if job.VersionID != "ver-1" {
		t.Errorf("VersionID: got %q, want %q", job.VersionID, "ver-1")
	}
	mock.AssertExpectations()
}

func TestJobs_CreateJob_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.CreateJob(context.Background(), "ver-1", "doc-1", "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestJobs_GetJob(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Job{
		{ID: "job-1", Status: models.JobPending, VersionID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", CreatedAt: ts, UpdatedAt: ts},
	})

	job, err := store.GetJob(context.Background(), "job-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("ID: got %q, want %q", job.ID, "job-1")
	}
	mock.AssertExpectations()
}

func TestJobs_GetJob_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetJob(context.Background(), "job-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestJobs_GetJobByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Job{
		{ID: "job-1", Status: models.JobProcessing, VersionID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", CreatedAt: ts, UpdatedAt: ts},
	})

	job, err := store.GetJobByTenant(context.Background(), "job-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("ID: got %q, want %q", job.ID, "job-1")
	}
	if job.TenantID != "t-1" {
		t.Errorf("TenantID: got %q, want %q", job.TenantID, "t-1")
	}
	mock.AssertExpectations()
}

func TestJobs_GetJobByTenant_NotFound(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	mock.ExpectQuery().WithRows([]models.Job{})

	_, err := store.GetJobByTenant(context.Background(), "job-1", "wrong-tenant")
	if err == nil {
		t.Fatal("expected error for tenant mismatch")
	}
	mock.AssertExpectations()
}

func TestJobs_GetJobByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.GetJobByTenant(context.Background(), "job-1", "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestJobs_UpdateJobStatus(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Modify().Exec() uses UPDATE...RETURNING on Postgres, so it's a query.
	mock.ExpectQuery().WithRows([]models.Job{
		{ID: "job-1", Status: models.JobCompleted, VersionID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", CreatedAt: ts, UpdatedAt: ts},
	})

	err := store.UpdateJobStatus(context.Background(), "job-1", models.JobCompleted, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestJobs_UpdateJobStatus_WithError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Job{
		{ID: "job-1", Status: models.JobFailed, VersionID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", CreatedAt: ts, UpdatedAt: ts},
	})

	errMsg := "extraction failed"
	err := store.UpdateJobStatus(context.Background(), "job-1", models.JobFailed, &errMsg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestJobs_UpdateJobStatus_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestJobs(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	err := store.UpdateJobStatus(context.Background(), "job-1", models.JobProcessing, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
