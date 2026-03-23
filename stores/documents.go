package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// Documents provides database access for documents.
type Documents struct {
	*sum.Database[models.Document]
}

// NewDocuments creates a new documents store.
func NewDocuments(db *sqlx.DB, renderer astql.Renderer) (*Documents, error) {
	database, err := sum.NewDatabase[models.Document](db, "documents", renderer)
	if err != nil {
		return nil, err
	}
	return &Documents{Database: database}, nil
}
