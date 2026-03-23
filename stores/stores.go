// Package stores provides data access layer implementations for all argus domain entities.
package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/grub"
)

// Stores aggregates all data store instances for the application.
type Stores struct {
	Tenants              *Tenants
	Providers            *Providers
	WatchedPaths         *WatchedPaths
	Documents            *Documents
	DocumentVersions     *DocumentVersions
	DocumentVersionSearch *DocumentVersionSearch
}

// New initializes all stores and returns an aggregated Stores instance.
func New(db *sqlx.DB, renderer astql.Renderer, searchProvider grub.SearchProvider) (*Stores, error) {
	tenants, err := NewTenants(db, renderer)
	if err != nil {
		return nil, err
	}

	providers, err := NewProviders(db, renderer)
	if err != nil {
		return nil, err
	}

	watchedPaths, err := NewWatchedPaths(db, renderer)
	if err != nil {
		return nil, err
	}

	documents, err := NewDocuments(db, renderer)
	if err != nil {
		return nil, err
	}

	documentVersions, err := NewDocumentVersions(db, renderer)
	if err != nil {
		return nil, err
	}

	documentVersionSearch, err := NewDocumentVersionSearch(searchProvider)
	if err != nil {
		return nil, err
	}

	return &Stores{
		Tenants:               tenants,
		Providers:             providers,
		WatchedPaths:          watchedPaths,
		Documents:             documents,
		DocumentVersions:      documentVersions,
		DocumentVersionSearch: documentVersionSearch,
	}, nil
}
