// Package stores provides data access layer implementations for all argus domain entities.
package stores

import (
	"github.com/jmoiron/sqlx"
	"github.com/zoobz-io/astql"
	"github.com/zoobz-io/grub"
)

// Stores aggregates all data store instances for the application.
type Stores struct {
	Tenants               *Tenants
	Providers             *Providers
	WatchedPaths          *WatchedPaths
	Documents             *Documents
	DocumentVersions      *DocumentVersions
	DocumentVersionSearch *DocumentVersionSearch
	Jobs                  *Jobs
	Topics                *Topics
	Tags                  *Tags
	Users                 *Users
	Subscriptions      *Subscriptions
	AdminSubscriptions *AdminSubscriptions
	Notifications      *Notifications
}

// New initializes all stores and returns an aggregated Stores instance.
func New(db *sqlx.DB, renderer astql.Renderer, bucket grub.BucketProvider, searchProvider grub.SearchProvider) *Stores {
	subs := NewSubscriptions(db, renderer)
	return &Stores{
		Tenants:               NewTenants(db, renderer),
		Providers:             NewProviders(db, renderer),
		WatchedPaths:          NewWatchedPaths(db, renderer),
		Documents:             NewDocuments(db, renderer),
		DocumentVersions:      NewDocumentVersions(db, renderer, bucket),
		DocumentVersionSearch: NewDocumentVersionSearch(searchProvider),
		Jobs:                  NewJobs(db, renderer),
		Topics:                NewTopics(db, renderer),
		Tags:                  NewTags(db, renderer),
		Users:                 NewUsers(db, renderer),
		Subscriptions:         subs,
		AdminSubscriptions:    &AdminSubscriptions{subs},
		Notifications:         NewNotifications(searchProvider),
	}
}
