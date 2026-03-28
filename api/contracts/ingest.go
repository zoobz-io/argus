package contracts

import "context"

// Ingest defines the contract for running the ingestion pipeline.
type Ingest interface {
	// Ingest processes a document version through the ingestion pipeline.
	// jobID is the pre-created job identifier.
	Ingest(ctx context.Context, jobID, versionID string) error
}
