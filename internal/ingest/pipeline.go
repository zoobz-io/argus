// Package ingest provides the document ingestion pipeline.
//
// The pipeline processes documents through extraction, summarization, embedding,
// and indexing stages. Each stage is a pipz processor that can be composed into
// different pipeline configurations.
package ingest

import (
	"context"
	"fmt"
	"log"

	"github.com/zoobz-io/argus/models"
)

// DocumentStore defines the store operations needed by the ingestion pipeline.
type DocumentStore interface {
	GetDocumentVersion(ctx context.Context, id int64) (*models.DocumentVersion, error)
}

// SearchIndex defines the search index operations needed by the ingestion pipeline.
type SearchIndex interface {
	IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error
}

// Pipeline orchestrates document ingestion through extraction, summarization,
// embedding, and indexing stages.
type Pipeline struct {
	documents DocumentStore
	search    SearchIndex
}

// New creates a new ingestion pipeline.
func New(documents DocumentStore, search SearchIndex) *Pipeline {
	return &Pipeline{
		documents: documents,
		search:    search,
	}
}

// Ingest processes a document version through the full ingestion pipeline.
// This is a stub — stages will be implemented as pipz processors.
func (p *Pipeline) Ingest(ctx context.Context, versionID int64) error {
	version, err := p.documents.GetDocumentVersion(ctx, versionID)
	if err != nil {
		return fmt.Errorf("fetching document version: %w", err)
	}

	log.Printf("ingest: starting pipeline for version_id=%d document_id=%d", version.ID, version.DocumentID)

	// Stage 1: Extract text content from the stored object.
	// TODO: Implement using pipz processor + OCR gRPC client.
	log.Printf("ingest: [stub] extract text from object_key=%s", version.ObjectKey)

	// Stage 2: Summarize extracted content.
	// TODO: Implement using pipz processor + LLM via zyn.
	log.Printf("ingest: [stub] summarize content")

	// Stage 3: Generate embedding vector.
	// TODO: Implement using pipz processor + vex embedding provider.
	log.Printf("ingest: [stub] generate embedding")

	// Stage 4: Index the document version in OpenSearch.
	// TODO: Build DocumentVersionIndex from extracted data and index it.
	log.Printf("ingest: [stub] index version in search")

	log.Printf("ingest: pipeline complete for version_id=%d", version.ID)
	return nil
}
