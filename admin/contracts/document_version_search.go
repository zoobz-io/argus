// Package contracts defines interface boundaries for the admin API surface.
package contracts

import (
	"context"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch defines admin operations for the document version search index.
type DocumentVersionSearch interface {
	// Get retrieves a document version from the search index by ID.
	Get(ctx context.Context, id string) (*grub.Document[models.DocumentVersionIndex], error)
	// Execute runs a search query against the document version index.
	Execute(ctx context.Context, search *lucene.Search) (*grub.SearchResult[models.DocumentVersionIndex], error)
	// Count returns the number of documents matching the query.
	Count(ctx context.Context, query lucene.Query) (int64, error)
	// Index indexes a document version in the search index.
	Index(ctx context.Context, id string, doc *models.DocumentVersionIndex) error
	// Delete removes a document version from the search index.
	Delete(ctx context.Context, id string) error
}
