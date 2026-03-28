//go:build testing

package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

// --- mocks ---

type mockIngestVersions struct {
	OnGetDocumentVersion func(ctx context.Context, id string) (*models.DocumentVersion, error)
	OnGetVersionContent  func(ctx context.Context, objectKey string) ([]byte, error)
}

func (m *mockIngestVersions) GetDocumentVersion(ctx context.Context, id string) (*models.DocumentVersion, error) {
	return m.OnGetDocumentVersion(ctx, id)
}

func (m *mockIngestVersions) GetVersionContent(ctx context.Context, objectKey string) ([]byte, error) {
	if m.OnGetVersionContent != nil {
		return m.OnGetVersionContent(ctx, objectKey)
	}
	return nil, nil
}

type mockIngestDocuments struct {
	OnGetDocument func(ctx context.Context, id string) (*models.Document, error)
}

func (m *mockIngestDocuments) GetDocument(ctx context.Context, id string) (*models.Document, error) {
	return m.OnGetDocument(ctx, id)
}

// --- helpers ---

func defaultVersion() *models.DocumentVersion {
	return &models.DocumentVersion{
		ID:         "ver-1",
		DocumentID: "doc-1",
		TenantID:   "tenant-1",
	}
}

func defaultDocument() *models.Document {
	return &models.Document{
		ID:        "doc-1",
		Name:      "report.pdf",
		MimeType:  "application/pdf",
		ObjectKey: "objects/doc-1/ver-1",
	}
}

func defaultJob() *models.Job {
	return &models.Job{
		ID:     "job-1",
		Status: models.JobPending,
	}
}

// setupPipelineRegistry registers the 3 contracts used by Pipeline.Ingest() before
// it calls p.sequence.Process(). For full pipeline tests, stage-level contracts
// must also be registered.
func setupPipelineRegistry(t *testing.T, opts ...func(sum.Key)) context.Context {
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

func withVersions(v intcontracts.IngestVersions) func(sum.Key) {
	return func(k sum.Key) { sum.Register[intcontracts.IngestVersions](k, v) }
}

func withDocuments(d intcontracts.IngestDocuments) func(sum.Key) {
	return func(k sum.Key) { sum.Register[intcontracts.IngestDocuments](k, d) }
}

func withJobs(j intcontracts.IngestJobs) func(sum.Key) {
	return func(k sum.Key) { sum.Register[intcontracts.IngestJobs](k, j) }
}

// --- tests ---

func TestNewSynapseAnalyzer(t *testing.T) {
	// NewSynapseAnalyzer takes a real *zyn.ExtractionSynapse — verify it wraps correctly.
	analyzer := NewSynapseAnalyzer(nil)
	if analyzer == nil {
		t.Fatal("expected non-nil analyzer")
	}
}

func TestPipeline_Ingest_FullSuccess(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return defaultVersion(), nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return defaultDocument(), nil
		},
	}
	jobs := &mockIngestJobs{
		OnGetJob: func(_ context.Context, id string) (*models.Job, error) {
			return defaultJob(), nil
		},
		OnUpdateJobStatus: func(_ context.Context, _ string, _ models.JobStatus, _ *string) error {
			return nil
		},
	}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))

	// Build a pipeline with a no-op sequence to bypass stage dependencies.
	p := &Pipeline{
		sequence: pipz.NewSequence[*DocumentContext](PipelineID),
	}

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipeline_Ingest_CompletedStatusUpdateError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return defaultVersion(), nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return defaultDocument(), nil
		},
	}
	jobs := &mockIngestJobs{
		OnGetJob: func(_ context.Context, _ string) (*models.Job, error) {
			return defaultJob(), nil
		},
		OnUpdateJobStatus: func(_ context.Context, _ string, status models.JobStatus, _ *string) error {
			if status == models.JobCompleted {
				return errors.New("db write failed")
			}
			return nil
		},
	}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	p := &Pipeline{
		sequence: pipz.NewSequence[*DocumentContext](PipelineID),
	}

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "updating job to completed") {
		t.Errorf("error should mention updating job to completed, got %q", err.Error())
	}
}

func TestPipeline_Ingest_VersionFetchError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return nil, errors.New("version not found")
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return defaultDocument(), nil
		},
	}
	jobs := &mockIngestJobs{}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	p := New()

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching document version") {
		t.Errorf("error should mention fetching version, got %q", err.Error())
	}
}

func TestPipeline_Ingest_DocumentFetchError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return defaultVersion(), nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return nil, errors.New("document not found")
		},
	}
	jobs := &mockIngestJobs{}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	p := New()

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching document") {
		t.Errorf("error should mention fetching document, got %q", err.Error())
	}
}

func TestPipeline_Ingest_GetJobError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return defaultVersion(), nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return defaultDocument(), nil
		},
	}
	jobs := &mockIngestJobs{
		OnGetJob: func(_ context.Context, _ string) (*models.Job, error) {
			return nil, errors.New("job not found")
		},
	}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	p := New()

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching job") {
		t.Errorf("error should mention fetching job, got %q", err.Error())
	}
}

func TestPipeline_Ingest_JobStatusUpdateError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return defaultVersion(), nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return defaultDocument(), nil
		},
	}
	jobs := &mockIngestJobs{
		OnGetJob: func(_ context.Context, _ string) (*models.Job, error) {
			return defaultJob(), nil
		},
		OnUpdateJobStatus: func(_ context.Context, _ string, status models.JobStatus, _ *string) error {
			if status == models.JobProcessing {
				return errors.New("status update failed")
			}
			return nil
		},
	}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	p := New()

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "setting job to processing") {
		t.Errorf("error should mention setting job to processing, got %q", err.Error())
	}
}

func TestPipeline_Ingest_PipelineProcessError(t *testing.T) {
	var failedJobID string
	var failedStatus models.JobStatus
	var failedErr *string

	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return defaultVersion(), nil
		},
		OnGetVersionContent: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("raw bytes"), nil
		},
	}
	documents := &mockIngestDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return defaultDocument(), nil
		},
	}
	jobs := &mockIngestJobs{
		OnGetJob: func(_ context.Context, _ string) (*models.Job, error) {
			return defaultJob(), nil
		},
		OnUpdateJobStatus: func(_ context.Context, id string, status models.JobStatus, jobErr *string) error {
			failedJobID = id
			failedStatus = status
			failedErr = jobErr
			return nil
		},
	}

	ctx := setupPipelineRegistry(t, withVersions(versions), withDocuments(documents), withJobs(jobs))
	p := New()

	err := p.Ingest(ctx, "job-1", "ver-1")
	if err == nil {
		t.Fatal("expected error from pipeline failure")
	}
	if !strings.Contains(err.Error(), "pipeline failed") {
		t.Errorf("error should mention pipeline failed, got %q", err.Error())
	}
	if failedJobID != "job-1" {
		t.Errorf("failed job ID: got %q, want %q", failedJobID, "job-1")
	}
	if failedStatus != models.JobFailed {
		t.Errorf("failed status: got %q, want %q", failedStatus, models.JobFailed)
	}
	if failedErr == nil {
		t.Error("expected error message on failed job, got nil")
	}
}
