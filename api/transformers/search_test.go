//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestSearchResultToResponse(t *testing.T) {
	result := &models.SearchResult{
		Hits: []models.SearchHit{
			{VersionID: "v1", DocumentID: "d1", DocumentName: "report.pdf", MimeType: "application/pdf", Summary: "A report", Score: 0.95, CreatedAt: argustest.FixtureTime, UpdatedAt: argustest.FixtureTime2},
			{VersionID: "v2", DocumentID: "d2", DocumentName: "memo.txt", MimeType: "text/plain", Summary: "A memo", Score: 0.80, CreatedAt: argustest.FixtureTime, UpdatedAt: argustest.FixtureTime2},
		},
		Facets: models.SearchFacets{
			"mime_type": {{Value: "application/pdf", Count: 5}, {Value: "text/plain", Count: 3}},
		},
		Total: 42,
	}
	page := wire.Page{Size: 20, From: 0}

	resp := SearchResultToResponse(result, page)

	// Hits mapped correctly.
	if len(resp.Hits) != 2 {
		t.Fatalf("expected 2 hits, got %d", len(resp.Hits))
	}
	if resp.Hits[0].VersionID != "v1" || resp.Hits[0].Score != 0.95 || resp.Hits[0].Summary != "A report" {
		t.Errorf("hit 0 mismatch: %+v", resp.Hits[0])
	}
	if resp.Hits[1].DocumentName != "memo.txt" {
		t.Errorf("hit 1 mismatch: %+v", resp.Hits[1])
	}

	// Facets mapped correctly.
	if len(resp.Facets["mime_type"]) != 2 {
		t.Fatalf("expected 2 facet buckets, got %d", len(resp.Facets["mime_type"]))
	}
	if resp.Facets["mime_type"][0].Value != "application/pdf" || resp.Facets["mime_type"][0].Count != 5 {
		t.Errorf("facet mismatch: %+v", resp.Facets["mime_type"][0])
	}

	// Pagination.
	if resp.Total != 42 {
		t.Errorf("total mismatch: %d", resp.Total)
	}
	if resp.Page.Size != 20 || resp.Page.From != 0 {
		t.Errorf("page mismatch: %+v", resp.Page)
	}
}

func TestSearchResultToResponse_HasMore(t *testing.T) {
	result := &models.SearchResult{Total: 50, Hits: []models.SearchHit{}}
	page := wire.Page{Size: 20, From: 0}

	resp := SearchResultToResponse(result, page)
	if !resp.Page.HasMore {
		t.Error("expected HasMore=true when from+size < total")
	}
}

func TestSearchResultToResponse_NoMore(t *testing.T) {
	result := &models.SearchResult{Total: 15, Hits: []models.SearchHit{}}
	page := wire.Page{Size: 20, From: 0}

	resp := SearchResultToResponse(result, page)
	if resp.Page.HasMore {
		t.Error("expected HasMore=false when from+size >= total")
	}
}

func TestSearchResultToResponse_EmptyFacets(t *testing.T) {
	result := &models.SearchResult{Hits: []models.SearchHit{}, Facets: nil}
	page := wire.Page{Size: 10}

	resp := SearchResultToResponse(result, page)
	if len(resp.Facets) != 0 {
		t.Errorf("expected empty facets, got %d", len(resp.Facets))
	}
}
