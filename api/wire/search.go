package wire

import (
	"context"
	"time"

	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

var validSortFields = []string{"", "relevance", "created_at", "updated_at", "name"}
var validSortOrders = []string{"", "asc", "desc"}
var validMatchModes = []string{"", "all", "any"}

// Sort controls result ordering.
type Sort struct {
	Field string `json:"field" description:"Sort field: relevance, created_at, updated_at, name" example:"created_at"`
	Order string `json:"order" description:"Sort direction: asc, desc" example:"desc"`
}

// DateRange specifies an optional time range with after/before bounds.
type DateRange struct {
	After  *time.Time `json:"after" description:"Include documents after this date"`
	Before *time.Time `json:"before" description:"Include documents before this date"`
}

// Page controls offset-based pagination for search requests and responses.
type Page struct {
	Size    int  `json:"size" description:"Number of results per page" example:"20"`
	From    int  `json:"from" description:"Offset of first result" example:"0"`
	HasMore bool `json:"has_more,omitempty" description:"Whether more results exist beyond this page"`
}

// SearchRequest is the request body for searching document versions.
type SearchRequest struct {
	Dates    map[string]DateRange `json:"dates" description:"Date range filters keyed by date field name"`
	Facets   map[string][]string  `json:"facets" description:"Facet filters keyed by facet name"`
	Query    string               `json:"query" description:"Natural language search query" example:"documents about revenue growth"`
	Keywords string               `json:"keywords" description:"Keyword search with Lucene syntax support" example:"+quarterly -draft"`
	Match    string               `json:"match" description:"How to combine query and keywords: all (AND) or any (OR)" example:"any"`
	Sort     Sort                 `json:"sort" description:"Result ordering"`
	Page     Page                 `json:"page" description:"Pagination parameters"`
}

// Validate validates the request fields.
func (r *SearchRequest) Validate() error {
	return check.All(
		check.Str(r.Match, "match").OneOf(validMatchModes).V(),
		check.Str(r.Sort.Field, "sort.field").OneOf(validSortFields).V(),
		check.Str(r.Sort.Order, "sort.order").OneOf(validSortOrders).V(),
		check.Int(r.Page.Size, "page.size").Min(1).Max(100).V(),
		check.Int(r.Page.From, "page.from").Min(0).V(),
	).Err()
}

// Clone returns a copy of the request.
func (r SearchRequest) Clone() SearchRequest {
	c := r
	if r.Dates != nil {
		c.Dates = make(map[string]DateRange, len(r.Dates))
		for k, v := range r.Dates {
			dr := v
			if v.After != nil {
				t := *v.After
				dr.After = &t
			}
			if v.Before != nil {
				t := *v.Before
				dr.Before = &t
			}
			c.Dates[k] = dr
		}
	}
	if r.Facets != nil {
		c.Facets = make(map[string][]string, len(r.Facets))
		for k, v := range r.Facets {
			cp := make([]string, len(v))
			copy(cp, v)
			c.Facets[k] = cp
		}
	}
	return c
}

// SearchHit represents a single search result.
type SearchHit struct {
	CreatedAt    time.Time `json:"created_at" description:"Version creation timestamp"`
	UpdatedAt    time.Time `json:"updated_at" description:"Document last updated timestamp"`
	DocumentName string    `json:"document_name" description:"Document name"`
	MimeType     string    `json:"mime_type" description:"MIME type"`
	Summary      string    `json:"summary" description:"AI-generated summary"`
	VersionID    string    `json:"version_id" description:"Document version ID"`
	DocumentID   string    `json:"document_id" description:"Document ID"`
	Score        float64   `json:"score" description:"Relevance score"`
}

// Clone returns a copy of the hit.
func (h SearchHit) Clone() SearchHit {
	return h
}

// SearchFacets maps facet names to their bucket counts.
type SearchFacets map[string][]FacetCount

// FacetCount represents a single facet bucket.
type FacetCount struct {
	Value string `json:"value" description:"Facet value"`
	Count int64  `json:"count" description:"Number of matching documents"`
}

// SearchResponse is the public API response for a search query.
type SearchResponse struct {
	Facets SearchFacets `json:"facets" description:"Aggregated facet counts"`
	Hits   []SearchHit  `json:"hits" description:"Search results"`
	Page   Page         `json:"page" description:"Pagination state"`
	Total  int64        `json:"total" description:"Total matching documents"`
}

// OnSend applies boundary masking before the response is marshaled.
func (s *SearchResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[SearchResponse]](ctx)
	masked, err := b.Send(ctx, *s)
	if err != nil {
		return err
	}
	*s = masked
	return nil
}

// Clone returns a deep copy of the response.
func (s SearchResponse) Clone() SearchResponse {
	c := s
	if s.Hits != nil {
		c.Hits = make([]SearchHit, len(s.Hits))
		copy(c.Hits, s.Hits)
	}
	if s.Facets != nil {
		c.Facets = make(SearchFacets, len(s.Facets))
		for name, buckets := range s.Facets {
			cp := make([]FacetCount, len(buckets))
			copy(cp, buckets)
			c.Facets[name] = cp
		}
	}
	return c
}
