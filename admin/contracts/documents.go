package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Documents defines admin operations for documents.
type Documents interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, id int64) (*models.Document, error)
	// DeleteDocument removes a document.
	DeleteDocument(ctx context.Context, id int64) error
	// ListDocuments retrieves a paginated list of all documents.
	ListDocuments(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Document], error)
	// ListDocumentsByTenant retrieves documents for a specific tenant.
	ListDocumentsByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Document], error)
}
