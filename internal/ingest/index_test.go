//go:build testing

package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/vex"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

// --- mocks ---

type mockIngestSearch struct {
	OnIndexVersion func(ctx context.Context, version *models.DocumentVersionIndex) error
}

func (m *mockIngestSearch) IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error {
	return m.OnIndexVersion(ctx, version)
}

type mockIngestJobs struct {
	OnCreateJob       func(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error)
	OnUpdateJobStatus func(ctx context.Context, id string, status models.JobStatus, jobErr *string) error
}

func (m *mockIngestJobs) CreateJob(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error) {
	if m.OnCreateJob != nil {
		return m.OnCreateJob(ctx, versionID, documentID, tenantID)
	}
	return &models.Job{ID: "job-default"}, nil
}

func (m *mockIngestJobs) UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, jobErr *string) error {
	if m.OnUpdateJobStatus != nil {
		return m.OnUpdateJobStatus(ctx, id, status, jobErr)
	}
	return nil
}

// --- helpers ---

func setupIndexRegistry(t *testing.T, search intcontracts.IngestSearch, jobs intcontracts.IngestJobs) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	sum.Register[intcontracts.IngestSearch](k, search)
	sum.Register[intcontracts.IngestJobs](k, jobs)
	sum.Freeze(k)
	return context.Background()
}

func testDocumentContext() *DocumentContext {
	return &DocumentContext{
		Version: &models.DocumentVersion{
			ID:         "ver-1",
			DocumentID: "doc-1",
			TenantID:   "tenant-1",
			CreatedAt:  time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
		},
		Document: &models.Document{
			ID:         "doc-1",
			ProviderID: "provider-1",
			Name:       "report.pdf",
			MimeType:   "application/pdf",
			UpdatedAt:  time.Date(2025, 6, 2, 0, 0, 0, 0, time.UTC),
		},
		Job: &models.Job{
			ID:     "job-1",
			Status: models.JobProcessing,
		},
		Content:   "extracted text",
		Summary:   "a summary",
		Language:  "en",
		Topics:    []string{"Finance"},
		Tags:      []string{"Important"},
		Embedding: vex.Vector{0.1, 0.2},
	}
}

// --- tests ---

func TestIndexStage_HappyPath(t *testing.T) {
	var indexed *models.DocumentVersionIndex
	var jobStatus models.JobStatus

	search := &mockIngestSearch{
		OnIndexVersion: func(_ context.Context, version *models.DocumentVersionIndex) error {
			indexed = version
			return nil
		},
	}
	jobs := &mockIngestJobs{
		OnUpdateJobStatus: func(_ context.Context, id string, status models.JobStatus, jobErr *string) error {
			if id != "job-1" {
				t.Errorf("job ID: got %q, want %q", id, "job-1")
			}
			jobStatus = status
			if jobErr != nil {
				t.Errorf("jobErr: expected nil, got %q", *jobErr)
			}
			return nil
		},
	}

	ctx := setupIndexRegistry(t, search, jobs)
	stage := newIndexStage()
	dc := testDocumentContext()

	_, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify field mapping.
	if indexed == nil {
		t.Fatal("IndexVersion was not called")
	}
	if indexed.VersionID != "ver-1" {
		t.Errorf("VersionID: got %q, want %q", indexed.VersionID, "ver-1")
	}
	if indexed.DocumentID != "doc-1" {
		t.Errorf("DocumentID: got %q, want %q", indexed.DocumentID, "doc-1")
	}
	if indexed.TenantID != "tenant-1" {
		t.Errorf("TenantID: got %q, want %q", indexed.TenantID, "tenant-1")
	}
	if indexed.ProviderID != "provider-1" {
		t.Errorf("ProviderID: got %q, want %q", indexed.ProviderID, "provider-1")
	}
	if indexed.DocumentName != "report.pdf" {
		t.Errorf("DocumentName: got %q, want %q", indexed.DocumentName, "report.pdf")
	}
	if indexed.MimeType != "application/pdf" {
		t.Errorf("MimeType: got %q, want %q", indexed.MimeType, "application/pdf")
	}
	if indexed.Content != "extracted text" {
		t.Errorf("Content: got %q, want %q", indexed.Content, "extracted text")
	}
	if indexed.Summary != "a summary" {
		t.Errorf("Summary: got %q, want %q", indexed.Summary, "a summary")
	}
	if indexed.Language != "en" {
		t.Errorf("Language: got %q, want %q", indexed.Language, "en")
	}
	if len(indexed.Topics) != 1 || indexed.Topics[0] != "Finance" {
		t.Errorf("Topics: got %v, want [Finance]", indexed.Topics)
	}
	if len(indexed.Tags) != 1 || indexed.Tags[0] != "Important" {
		t.Errorf("Tags: got %v, want [Important]", indexed.Tags)
	}
	if len(indexed.Embedding) != 2 {
		t.Errorf("Embedding len: got %d, want 2", len(indexed.Embedding))
	}
	if !indexed.CreatedAt.Equal(dc.Version.CreatedAt) {
		t.Errorf("CreatedAt: got %v, want %v", indexed.CreatedAt, dc.Version.CreatedAt)
	}
	if !indexed.UpdatedAt.Equal(dc.Document.UpdatedAt) {
		t.Errorf("UpdatedAt: got %v, want %v", indexed.UpdatedAt, dc.Document.UpdatedAt)
	}

	// Verify job completed.
	if jobStatus != models.JobCompleted {
		t.Errorf("job status: got %q, want %q", jobStatus, models.JobCompleted)
	}
}

func TestIndexStage_IndexError(t *testing.T) {
	search := &mockIngestSearch{
		OnIndexVersion: func(_ context.Context, _ *models.DocumentVersionIndex) error {
			return errors.New("index failure")
		},
	}
	jobs := &mockIngestJobs{}

	ctx := setupIndexRegistry(t, search, jobs)
	stage := newIndexStage()
	dc := testDocumentContext()

	_, err := stage.Process(ctx, dc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "indexing version: index failure") {
		t.Errorf("error should contain 'indexing version: index failure', got %q", err.Error())
	}
}

func TestIndexStage_JobUpdateError(t *testing.T) {
	search := &mockIngestSearch{
		OnIndexVersion: func(_ context.Context, _ *models.DocumentVersionIndex) error {
			return nil
		},
	}
	jobs := &mockIngestJobs{
		OnUpdateJobStatus: func(_ context.Context, _ string, _ models.JobStatus, _ *string) error {
			return errors.New("job update failure")
		},
	}

	ctx := setupIndexRegistry(t, search, jobs)
	stage := newIndexStage()
	dc := testDocumentContext()

	_, err := stage.Process(ctx, dc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "updating job status: job update failure") {
		t.Errorf("error should contain 'updating job status: job update failure', got %q", err.Error())
	}
}
