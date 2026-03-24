package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersions defines the contract for document version operations on the public API surface.
type DocumentVersions interface {
	// GetDocumentVersion retrieves a document version by ID.
	GetDocumentVersion(ctx context.Context, id int64) (*models.DocumentVersion, error)
	// ListVersionsByDocument retrieves versions for a document using cursor pagination.
	ListVersionsByDocument(ctx context.Context, documentID int64, page models.CursorPage) (*models.CursorResult[models.DocumentVersion], error)
}
