package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// Providers provides database access for providers.
type Providers struct {
	*sum.Database[models.Provider]
}

// NewProviders creates a new providers store.
func NewProviders(db *sqlx.DB, renderer astql.Renderer) (*Providers, error) {
	database, err := sum.NewDatabase[models.Provider](db, "providers", renderer)
	if err != nil {
		return nil, err
	}
	return &Providers{Database: database}, nil
}
