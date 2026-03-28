package events

import (
	"github.com/zoobz-io/capitan"
)

// Ingest pipeline signals.
var (
	IngestStarted          = capitan.NewSignal("argus.ingest.started", "Ingestion pipeline started")
	IngestExtracted        = capitan.NewSignal("argus.ingest.extracted", "Text content extracted from document")
	IngestSummarized       = capitan.NewSignal("argus.ingest.summarized", "Content summary generated")
	IngestSummarizeFailed  = capitan.NewSignal("argus.ingest.summarize.failed", "Content summarization failed")
	IngestEmbedded         = capitan.NewSignal("argus.ingest.embedded", "Embedding vector generated")
	IngestEmbedFailed      = capitan.NewSignal("argus.ingest.embed.failed", "Embedding generation failed")
	IngestIndexed          = capitan.NewSignal("argus.ingest.indexed", "Document version indexed in search")
	IngestClassifyUnavailable = capitan.NewSignal("argus.ingest.classify.unavailable", "Injection classifier unavailable, failing open")
	IngestClassifyRejected    = capitan.NewSignal("argus.ingest.classify.rejected", "Content flagged as prompt injection")
	IngestCompleted           = capitan.NewSignal("argus.ingest.completed", "Ingestion pipeline completed")
	IngestFailed              = capitan.NewSignal("argus.ingest.failed", "Ingestion pipeline failed")
)

// Ingest field keys for signal emission.
var (
	IngestVersionIDKey  = capitan.NewStringKey("version_id")
	IngestDocumentIDKey = capitan.NewStringKey("document_id")
	IngestTenantIDKey   = capitan.NewStringKey("tenant_id")
	IngestMimeTypeKey   = capitan.NewStringKey("mime_type")
	IngestErrorKey      = capitan.NewErrorKey("error")
)
