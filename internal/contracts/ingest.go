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
	GetDocumentVersion(ctx context.Context, id string) (*models.DocumentVersion, error)
	// GetVersionContent retrieves the raw bytes for a document version from object storage.
	GetVersionContent(ctx context.Context, objectKey string) ([]byte, error)
}

// IngestJobs defines the job tracking operations needed by the ingestion pipeline.
type IngestJobs interface {
	// CreateJob creates a new ingestion job.
	CreateJob(ctx context.Context, versionID, documentID, tenantID string) (*models.Job, error)
	// UpdateJobStatus sets the status and optional error for a job.
	UpdateJobStatus(ctx context.Context, id string, status models.JobStatus, jobErr *string) error
}

// IngestDocuments defines the document operations needed by the ingestion pipeline.
type IngestDocuments interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, id string) (*models.Document, error)
}

// IngestTopics defines the topic operations needed by the ingestion pipeline.
type IngestTopics interface {
	// ListTopicsByTenant retrieves all topics for a tenant.
	ListTopicsByTenant(ctx context.Context, tenantID string) ([]*models.Topic, error)
}

// IngestTags defines the tag operations needed by the ingestion pipeline.
type IngestTags interface {
	// ListTagsByTenant retrieves all tags for a tenant.
	ListTagsByTenant(ctx context.Context, tenantID string) ([]*models.Tag, error)
}

// IngestSearch defines the search index operations needed by the ingestion pipeline.
type IngestSearch interface {
	// IndexVersion indexes a document version in the search index.
	IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error
}

// OCR defines the text extraction operations needed by the ingestion pipeline.
// Satisfied by proto.OCRServiceClient.
type OCR = proto.OCRServiceClient

// Converter defines document format conversion operations needed by the ingestion pipeline.
// Satisfied by proto.ConvertServiceClient.
type Converter = proto.ConvertServiceClient

// Classifier defines prompt injection classification operations.
// Satisfied by proto.ClassifyServiceClient.
type Classifier = proto.ClassifyServiceClient

// Analyzer defines the LLM analysis operations needed by the ingestion pipeline.
type Analyzer interface {
	// Analyze extracts structured metadata from document content using tenant-defined vocabularies.
	Analyze(ctx context.Context, content string, topics, tags []string) (*models.DocumentAnalysis, error)
}

// Embedder defines the vector embedding operations needed by the ingestion pipeline.
type Embedder interface {
	// Embed generates a vector embedding for the given text.
	Embed(ctx context.Context, text string) (vex.Vector, error)
}

