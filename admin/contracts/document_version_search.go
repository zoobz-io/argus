// Package contracts defines interface boundaries for the admin API surface.
package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch defines admin operations for the document version search index.
type DocumentVersionSearch interface {
	// Search performs a full-text, semantic, or hybrid search across document versions.
	Search(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error)
	// GetDocumentEmbedding retrieves the embedding vector for a document.
	GetDocumentEmbedding(ctx context.Context, documentID string) ([]float32, error)
	// IndexVersion indexes a document version in the search index.
	IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error
	// DeleteDocument removes a document from the search index.
	DeleteDocument(ctx context.Context, documentID string) error
}
