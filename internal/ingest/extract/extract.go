// Package extract provides format-specific text extraction functions
// for the document ingestion pipeline.
//
// Each extractor takes raw document bytes and returns extracted text content.
// Extractors are registered by MIME type and composed into a pipz.Switch
// by the parent ingest package.
package extract

import "context"

// Func is a format-specific text extraction function.
// It receives raw document bytes and returns the extracted text content.
type Func func(ctx context.Context, data []byte) (string, error)
