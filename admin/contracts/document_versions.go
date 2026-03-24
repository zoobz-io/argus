package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersions defines admin operations for document versions.
type DocumentVersions interface {
	// GetDocumentVersion retrieves a document version by ID.
	GetDocumentVersion(ctx context.Context, id int64) (*models.DocumentVersion, error)
	// DeleteDocumentVersion removes a document version.
	DeleteDocumentVersion(ctx context.Context, id int64) error
	// ListDocumentVersions retrieves a paginated list of all document versions.
	ListDocumentVersions(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.DocumentVersion], error)
	// ListVersionsByDocument retrieves versions for a specific document.
	ListVersionsByDocument(ctx context.Context, documentID int64, page models.CursorPage) (*models.CursorResult[models.DocumentVersion], error)
}
