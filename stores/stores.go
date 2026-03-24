// Package stores provides data access layer implementations for all argus domain entities.
package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/argus/models"
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
func New(db *sqlx.DB, renderer astql.Renderer, bucket grub.BucketProvider, searchProvider grub.SearchProvider) (*Stores, error) {
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

	documentVersions, err := NewDocumentVersions(db, renderer, bucket)
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

// idGetter is a constraint for models with an ID field.
type idGetter interface {
	GetID() int64
}

// cursorResult builds a CursorResult from a fetched slice, where limit+1 items
// were requested to determine HasMore.
func cursorResult[T idGetter](items []*T, limit int) *models.CursorResult[T] {
	hasMore := len(items) > limit
	if hasMore {
		items = items[:limit]
	}
	var cursor *int64
	if len(items) > 0 {
		id := (*items[len(items)-1]).GetID()
		cursor = &id
	}
	return &models.CursorResult[T]{
		Items:   items,
		Cursor:  cursor,
		HasMore: hasMore,
	}
}
