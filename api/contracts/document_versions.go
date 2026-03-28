package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersions defines the contract for document version operations on the public API surface.
type DocumentVersions interface {
	// GetDocumentVersion retrieves a document version by ID.
	GetDocumentVersion(ctx context.Context, id string) (*models.DocumentVersion, error)
	// ListVersionsByDocument retrieves versions for a document using cursor pagination.
	ListVersionsByDocument(ctx context.Context, documentID string, page models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error)
}
