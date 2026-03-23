package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// Tenants provides database access for tenants.
type Tenants struct {
	*sum.Database[models.Tenant]
}

// NewTenants creates a new tenants store.
func NewTenants(db *sqlx.DB, renderer astql.Renderer) (*Tenants, error) {
	database, err := sum.NewDatabase[models.Tenant](db, "tenants", renderer)
	if err != nil {
		return nil, err
	}
	return &Tenants{Database: database}, nil
}
