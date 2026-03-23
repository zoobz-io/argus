package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersions provides database access for document versions.
type DocumentVersions struct {
	*sum.Database[models.DocumentVersion]
}

// NewDocumentVersions creates a new document versions store.
func NewDocumentVersions(db *sqlx.DB, renderer astql.Renderer) (*DocumentVersions, error) {
	database, err := sum.NewDatabase[models.DocumentVersion](db, "document_versions", renderer)
	if err != nil {
		return nil, err
	}
	return &DocumentVersions{Database: database}, nil
}
