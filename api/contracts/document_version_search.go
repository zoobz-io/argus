// Package contracts defines interface boundaries for the public API surface.
package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch defines the contract for searching document versions.
type DocumentVersionSearch interface {
	// Search performs a full-text, semantic, or hybrid search across document versions.
	Search(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error)
	// GetDocumentEmbedding retrieves the embedding vector for a document.
	GetDocumentEmbedding(ctx context.Context, documentID string) ([]float32, error)
}
