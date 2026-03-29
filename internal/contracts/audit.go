package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// AuditIndexer defines indexing operations needed by the audit pipeline.
type AuditIndexer interface {
	// Index persists an audit entry to the search index.
	Index(ctx context.Context, entry *models.AuditEntry) error
}
