// Package ingest provides the document ingestion pipeline.
//
// The pipeline processes documents through extraction, summarization, embedding,
// and indexing stages. Each stage is a pipz processor that can be composed into
// different pipeline configurations.
package ingest

import (
	"context"
	"fmt"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

// Stage identities for pipeline introspection and resolution.
var (
	PipelineID      = pipz.NewIdentity("ingest-pipeline", "Document ingestion pipeline")
	ExtractID       = pipz.NewIdentity("extract", "Extract text content from stored object")
	FetchID         = pipz.NewIdentity("extract-fetch", "Fetch raw bytes from object storage")
	ExtractRouterID = pipz.NewIdentity("extract-router", "Route to format-specific extractor by MIME type")
	ExtractSignalID = pipz.NewIdentity("extract-signal", "Emit extraction completed signal")
	AnalyzeID       = pipz.NewIdentity("analyze", "Analyze content: summarize, classify topics/tags, detect language")
	EmbedID         = pipz.NewIdentity("embed", "Generate embedding vector for content")
	IndexID         = pipz.NewIdentity("index", "Index document version in search")
)

// Pipeline orchestrates document ingestion through extraction, analysis,
// embedding, and indexing stages.
type Pipeline struct {
	sequence *pipz.Sequence[*DocumentContext]
}

// New creates a new ingestion pipeline.
func New() *Pipeline {
	seq := pipz.NewSequence(
		PipelineID,
		newExtractStage(),
		newClassifyStage(),
		newAnalyzeStage(),
		newEmbedStage(),
		newIndexStage(),
	)

	return &Pipeline{sequence: seq}
}

// Ingest processes a document version through the full ingestion pipeline.
// jobID is the pre-created job identifier from the enqueuer.
func (p *Pipeline) Ingest(ctx context.Context, jobID, versionID string) error {
	versions := sum.MustUse[intcontracts.IngestVersions](ctx)
	documents := sum.MustUse[intcontracts.IngestDocuments](ctx)
	jobs := sum.MustUse[intcontracts.IngestJobs](ctx)

	version, err := versions.GetDocumentVersion(ctx, versionID)
	if err != nil {
		return fmt.Errorf("fetching document version: %w", err)
	}

	document, err := documents.GetDocument(ctx, version.DocumentID)
	if err != nil {
		return fmt.Errorf("fetching document: %w", err)
	}

	job, err := jobs.GetJob(ctx, jobID)
	if err != nil {
		return fmt.Errorf("fetching job: %w", err)
	}

	if statusErr := jobs.UpdateJobStatus(ctx, job.ID, models.JobProcessing, nil); statusErr != nil {
		return fmt.Errorf("setting job to processing: %w", statusErr)
	}

	capitan.Info(ctx, events.IngestStarted,
		events.IngestVersionIDKey.Field(version.ID),
		events.IngestDocumentIDKey.Field(version.DocumentID),
		events.IngestTenantIDKey.Field(version.TenantID),
		events.IngestJobIDKey.Field(job.ID),
		events.IngestMimeTypeKey.Field(document.MimeType),
	)

	dc := &DocumentContext{
		Version:  version,
		Document: document,
		Job:      job,
	}

	_, err = p.sequence.Process(ctx, dc)
	if err != nil {
		errMsg := err.Error()
		_ = jobs.UpdateJobStatus(ctx, job.ID, models.JobFailed, &errMsg)
		capitan.Error(ctx, events.IngestFailed,
			events.IngestVersionIDKey.Field(version.ID),
			events.IngestDocumentIDKey.Field(version.DocumentID),
			events.IngestTenantIDKey.Field(version.TenantID),
			events.IngestJobIDKey.Field(job.ID),
			events.IngestErrorKey.Field(err),
		)
		return fmt.Errorf("pipeline failed: %w", err)
	}

	if err := jobs.UpdateJobStatus(ctx, job.ID, models.JobCompleted, nil); err != nil {
		return fmt.Errorf("updating job to completed: %w", err)
	}

	capitan.Info(ctx, events.IngestCompleted,
		events.IngestVersionIDKey.Field(version.ID),
		events.IngestDocumentIDKey.Field(version.DocumentID),
		events.IngestTenantIDKey.Field(version.TenantID),
		events.IngestJobIDKey.Field(job.ID),
	)
	return nil
}
