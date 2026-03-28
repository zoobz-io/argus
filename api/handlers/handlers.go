package handlers

import "github.com/zoobz-io/rocco"

// All returns the complete set of public API endpoints.
func All() []rocco.Endpoint {
	return []rocco.Endpoint{
		// Providers
		listProviders,
		getProvider,
		createProvider,
		updateProvider,
		// Watched Paths
		listWatchedPaths,
		getWatchedPath,
		createWatchedPath,
		updateWatchedPath,
		// Documents
		listDocuments,
		getDocument,
		// Document Versions
		listDocumentVersions,
		getDocumentVersion,
		// Search
		searchDocuments,
		relatedDocuments,
		// Topics
		listTopics,
		getTopic,
		createTopic,
		updateTopic,
		// Tags
		listTags,
		getTag,
		createTag,
		updateTag,
		// Ingestion
		triggerIngest,
	}
}
