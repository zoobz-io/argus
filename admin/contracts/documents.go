package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Documents defines admin operations for documents.
type Documents interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, id string) (*models.Document, error)
	// DeleteDocument removes a document.
	DeleteDocument(ctx context.Context, id string) error
	// ListDocuments retrieves a paginated list of all documents.
	ListDocuments(ctx context.Context, page models.OffsetPage) (*models.OffsetResult[models.Document], error)
	// ListDocumentsByTenant retrieves documents for a specific tenant.
	ListDocumentsByTenant(ctx context.Context, tenantID string, page models.OffsetPage) (*models.OffsetResult[models.Document], error)
}
