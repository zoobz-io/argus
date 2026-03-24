package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Documents defines the contract for document operations on the public API surface.
type Documents interface {
	// GetDocument retrieves a document by ID.
	GetDocument(ctx context.Context, id int64) (*models.Document, error)
	// ListDocumentsByTenant retrieves documents for a tenant using cursor pagination.
	ListDocumentsByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Document], error)
}
