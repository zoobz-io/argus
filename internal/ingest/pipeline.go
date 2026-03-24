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
	SummarizeID     = pipz.NewIdentity("summarize", "Summarize extracted content via LLM")
	EmbedID         = pipz.NewIdentity("embed", "Generate embedding vector for content")
	IndexID         = pipz.NewIdentity("index", "Index document version in search")
)

// Pipeline orchestrates document ingestion through extraction, summarization,
// embedding, and indexing stages.
type Pipeline struct {
	sequence *pipz.Sequence[*DocumentContext]
}

// New creates a new ingestion pipeline.
func New() *Pipeline {
	seq := pipz.NewSequence(
		PipelineID,
		newExtractStage(),
		newSummarizeStage(),
		newEmbedStage(),
		newIndexStage(),
	)

	return &Pipeline{sequence: seq}
}

// Ingest processes a document version through the full ingestion pipeline.
func (p *Pipeline) Ingest(ctx context.Context, versionID int64) error {
	versions := sum.MustUse[intcontracts.IngestVersions](ctx)
	documents := sum.MustUse[intcontracts.IngestDocuments](ctx)

	version, err := versions.GetDocumentVersion(ctx, versionID)
	if err != nil {
		return fmt.Errorf("fetching document version: %w", err)
	}

	document, err := documents.GetDocument(ctx, version.DocumentID)
	if err != nil {
		return fmt.Errorf("fetching document: %w", err)
	}

	if statusErr := versions.UpdateExtractionStatus(ctx, versionID, models.ExtractionProcessing); statusErr != nil {
		return fmt.Errorf("setting status to processing: %w", statusErr)
	}

	capitan.Info(ctx, events.IngestStarted,
		events.IngestVersionIDKey.Field(version.ID),
		events.IngestDocumentIDKey.Field(version.DocumentID),
		events.IngestMimeTypeKey.Field(document.MimeType),
	)

	dc := &DocumentContext{
		Version:  version,
		Document: document,
	}

	_, err = p.sequence.Process(ctx, dc)
	if err != nil {
		_ = versions.UpdateExtractionStatus(ctx, versionID, models.ExtractionFailed)
		capitan.Error(ctx, events.IngestFailed,
			events.IngestVersionIDKey.Field(version.ID),
			events.IngestErrorKey.Field(err),
		)
		return fmt.Errorf("pipeline failed: %w", err)
	}

	capitan.Info(ctx, events.IngestCompleted,
		events.IngestVersionIDKey.Field(version.ID),
		events.IngestDocumentIDKey.Field(version.DocumentID),
	)
	return nil
}
