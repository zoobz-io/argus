// Package contracts defines interface boundaries for the admin API surface.
package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch defines admin operations for the document version search index.
type DocumentVersionSearch interface {
	// Search performs a full-text search across document version content.
	Search(ctx context.Context, query string, size int, from int) (*models.SearchResult, error)
	// IndexVersion indexes a document version in the search index.
	IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error
	// DeleteVersion removes a document version from the search index.
	DeleteVersion(ctx context.Context, versionID int64) error
}
