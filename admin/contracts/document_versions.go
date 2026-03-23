package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersions defines admin operations for document versions.
type DocumentVersions interface {
	// Get retrieves a document version by primary key.
	Get(ctx context.Context, key string) (*models.DocumentVersion, error)
	// Delete removes a document version.
	Delete(ctx context.Context, key string) error
}
