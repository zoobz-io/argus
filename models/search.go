package models

import "time"

// SearchHit represents a single search result from the document version index.
type SearchHit struct {
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DocumentName string
	MimeType     string
	Summary      string
	VersionID    string
	DocumentID   string
	Score        float64
}

// SearchResult holds the results of a document version search.
type SearchResult struct {
	Facets SearchFacets
	Hits   []SearchHit
	Total  int64
}
