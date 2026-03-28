package models

import "time"

// DateRange specifies an optional time range with after/before bounds.
type DateRange struct {
	After  *time.Time
	Before *time.Time
}

// MatchMode controls how multiple query types are combined.
type MatchMode string

// Match mode constants.
const (
	MatchAny MatchMode = "any"
	MatchAll MatchMode = "all"
)

// SearchParams holds all parameters for a document version search query.
type SearchParams struct {
	Dates        map[string]DateRange
	Facets       map[string][]string
	Sort         SearchSort
	Query        string
	TenantID     string
	Match        MatchMode
	QueryVectors [][]float32
	Size         int
	From         int
}

// SearchSort defines the sort field and direction.
type SearchSort struct {
	Field string
	Order string
}

// SearchFacets maps facet names to their bucket counts.
type SearchFacets map[string][]FacetCount

// FacetCount represents a single facet bucket.
type FacetCount struct {
	Value string
	Count int64
}
