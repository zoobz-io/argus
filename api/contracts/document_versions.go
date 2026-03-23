package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersions defines the contract for document version data access.
type DocumentVersions interface {
	// Get retrieves a document version by primary key.
	Get(ctx context.Context, key string) (*models.DocumentVersion, error)
}
