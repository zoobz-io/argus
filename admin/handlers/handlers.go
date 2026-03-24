// Package handlers provides HTTP endpoint handlers for the admin API.
package handlers

import "github.com/zoobz-io/rocco"

// All returns the complete set of admin API endpoints.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Tenants
		listTenants,
		getTenant,
		createTenant,
		updateTenant,
		deleteTenant,
		// Providers
		listAllProviders,
		getAdminProvider,
		createAdminProvider,
		deleteAdminProvider,
		// Watched Paths
		listAllWatchedPaths,
		getAdminWatchedPath,
		deleteAdminWatchedPath,
		// Documents
		listAllDocuments,
		getAdminDocument,
		deleteAdminDocument,
		// Document Versions
		listAllDocumentVersions,
		getAdminDocumentVersion,
		deleteAdminDocumentVersion,
	}
}
