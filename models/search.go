package models

// SearchHit represents a single search result from the document version index.
type SearchHit struct {
	DocumentName string
	MimeType     string
	Summary      string
	VersionID    int64
	DocumentID   int64
	Score        float64
}

// SearchResult holds the results of a document version search.
type SearchResult struct {
	Hits  []SearchHit
	Total int64
}
