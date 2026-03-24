// Package contracts defines interface boundaries for the public API surface.
package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch defines the contract for searching document versions.
type DocumentVersionSearch interface {
	// Search performs a full-text search across document version content.
	Search(ctx context.Context, query string, size int, from int) (*models.SearchResult, error)
}
