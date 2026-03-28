//go:build testing

package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

func setupEnqueueRegistry(t *testing.T, opts ...func(sum.Key)) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	for _, opt := range opts {
		opt(k)
	}
	sum.Freeze(k)
	return context.Background()
}

func TestEnqueue_Success(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, id string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: id, DocumentID: "doc-1", TenantID: "tenant-1"}, nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return &models.Document{ID: "doc-1"}, nil
		},
	}
	var createdJob *models.Job
	jobs := &mockIngestJobs{
		OnCreateJob: func(_ context.Context, versionID, documentID, tenantID string) (*models.Job, error) {
			createdJob = &models.Job{ID: "job-1", VersionID: versionID, DocumentID: documentID, TenantID: tenantID, Status: models.JobPending}
			return createdJob, nil
		},
	}

	ctx := setupEnqueueRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	e := NewEnqueuer()

	job, err := e.Enqueue(ctx, "ver-1", "tenant-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("job ID: got %q, want %q", job.ID, "job-1")
	}
	if job.Status != models.JobPending {
		t.Errorf("job status: got %q, want %q", job.Status, models.JobPending)
	}
}

func TestEnqueue_VersionNotFound(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return nil, errors.New("not found")
		},
	}
	documents := &mockIngestDocuments{}
	jobs := &mockIngestJobs{}

	ctx := setupEnqueueRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	e := NewEnqueuer()

	_, err := e.Enqueue(ctx, "ver-1", "tenant-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching document version") {
		t.Errorf("error should mention fetching document version, got %q", err.Error())
	}
}

func TestEnqueue_TenantMismatch(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-2"}, nil
		},
	}
	documents := &mockIngestDocuments{}
	jobs := &mockIngestJobs{}

	ctx := setupEnqueueRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	e := NewEnqueuer()

	_, err := e.Enqueue(ctx, "ver-1", "tenant-1")
	if err == nil {
		t.Fatal("expected error for tenant mismatch, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("error should say not found, got %q", err.Error())
	}
}

func TestEnqueue_DocumentNotFound(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"}, nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return nil, errors.New("not found")
		},
	}
	jobs := &mockIngestJobs{}

	ctx := setupEnqueueRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	e := NewEnqueuer()

	_, err := e.Enqueue(ctx, "ver-1", "tenant-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching document") {
		t.Errorf("error should mention fetching document, got %q", err.Error())
	}
}

func TestEnqueue_JobCreationError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"}, nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return &models.Document{ID: "doc-1"}, nil
		},
	}
	jobs := &mockIngestJobs{
		OnCreateJob: func(_ context.Context, _, _, _ string) (*models.Job, error) {
			return nil, errors.New("db error")
		},
	}

	ctx := setupEnqueueRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	e := NewEnqueuer()

	_, err := e.Enqueue(ctx, "ver-1", "tenant-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "creating job") {
		t.Errorf("error should mention creating job, got %q", err.Error())
	}
}
