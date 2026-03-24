// Package contracts defines interface boundaries for internal services.
package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/proto"
	"github.com/zoobz-io/vex"
)

// IngestVersions defines the document version operations needed by the ingestion pipeline.
type IngestVersions interface {
	// GetDocumentVersion retrieves a document version by ID.
	GetDocumentVersion(ctx context.Context, id int64) (*models.DocumentVersion, error)
	// GetVersionContent retrieves the raw bytes for a document version from object storage.
	GetVersionContent(ctx context.Context, objectKey string) ([]byte, error)
	// UpdateExtractionStatus updates the extraction status for a document version.
	UpdateExtractionStatus(ctx context.Context, id int64, status models.ExtractionStatus) error
}

// IngestDocuments defines the document operations needed by the ingestion pipeline.
type IngestDocuments interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, id int64) (*models.Document, error)
}

// IngestSearch defines the search index operations needed by the ingestion pipeline.
type IngestSearch interface {
	// IndexVersion indexes a document version in the search index.
	IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error
}

// OCR defines the text extraction operations needed by the ingestion pipeline.
// Satisfied by proto.OCRServiceClient.
type OCR = proto.OCRServiceClient

// Summarizer defines the LLM summarization operations needed by the ingestion pipeline.
type Summarizer interface {
	// Summarize generates a text summary of the given content.
	Summarize(ctx context.Context, content string) (string, error)
}

// Embedder defines the vector embedding operations needed by the ingestion pipeline.
type Embedder interface {
	// Embed generates a vector embedding for the given text.
	Embed(ctx context.Context, text string) (vex.Vector, error)
}

// TextExtractor defines the LLM-based text extraction operations needed by the ingestion pipeline.
type TextExtractor interface {
	// Extract cleans raw pre-processed document text into readable content via LLM.
	Extract(ctx context.Context, rawText string) (string, error)
}
