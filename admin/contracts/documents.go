package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Documents defines admin operations for documents.
type Documents interface {
	// Get retrieves a document by primary key.
	Get(ctx context.Context, key string) (*models.Document, error)
	// Delete removes a document.
	Delete(ctx context.Context, key string) error
}
