package contracts

import "context"

// Ingest defines the contract for triggering document ingestion.
type Ingest interface {
	// Ingest processes a document version through the ingestion pipeline.
	Ingest(ctx context.Context, versionID string) error
}
