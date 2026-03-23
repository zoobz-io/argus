package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Documents defines the contract for document data access.
type Documents interface {
	// Get retrieves a document by primary key.
	Get(ctx context.Context, key string) (*models.Document, error)
}
