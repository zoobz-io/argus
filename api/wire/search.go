package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// SearchRequest is the request body for searching document versions.
type SearchRequest struct {
	Query string `json:"query" description:"Search query text" example:"quarterly report"`
	Size  int    `json:"size" description:"Number of results" example:"20"`
	From  int    `json:"from" description:"Offset for pagination" example:"0"`
}

// Clone returns a copy of the request.
func (r SearchRequest) Clone() SearchRequest {
	return r
}

// SearchHit represents a single search result.
type SearchHit struct {
	DocumentName string  `json:"document_name" description:"Document name"`
	MimeType     string  `json:"mime_type" description:"MIME type"`
	Summary      string  `json:"summary" description:"AI-generated summary"`
	VersionID    int64   `json:"version_id" description:"Document version ID"`
	DocumentID   int64   `json:"document_id" description:"Document ID"`
	Score        float64 `json:"score" description:"Relevance score"`
}

// Clone returns a copy of the hit.
func (h SearchHit) Clone() SearchHit {
	return h
}

// SearchResponse is the public API response for a search query.
type SearchResponse struct {
	Hits  []SearchHit `json:"hits" description:"Search results"`
	Total int64       `json:"total" description:"Total matching documents"`
}

// OnSend applies boundary masking before the response is marshaled.
func (s *SearchResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[SearchResponse]](ctx)
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
	return c
}
