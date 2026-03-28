//go:build testing

package wire

import (
	"testing"
	"time"
)

func TestSearchRequest_Validate_Valid(t *testing.T) {
	tests := []struct {
		name string
		req  SearchRequest
	}{
		{
			name: "minimal valid request",
			req:  SearchRequest{Page: Page{Size: 20}},
		},
		{
			name: "all fields populated",
			req: SearchRequest{
				Query:    "revenue growth",
				Keywords: "+quarterly -draft",
				Match:    "all",
				Sort:     Sort{Field: "created_at", Order: "desc"},
				Page:     Page{Size: 50, From: 10},
			},
		},
		{
			name: "match mode any",
			req:  SearchRequest{Match: "any", Page: Page{Size: 1}},
		},
		{
			name: "empty match mode is valid",
			req:  SearchRequest{Match: "", Page: Page{Size: 1}},
		},
		{
			name: "all sort fields",
			req:  SearchRequest{Sort: Sort{Field: "relevance"}, Page: Page{Size: 1}},
		},
		{
			name: "sort by name asc",
			req:  SearchRequest{Sort: Sort{Field: "name", Order: "asc"}, Page: Page{Size: 1}},
		},
		{
			name: "page size at max",
			req:  SearchRequest{Page: Page{Size: 100}},
		},
		{
			name: "page size at min",
			req:  SearchRequest{Page: Page{Size: 1}},
		},
		{
			name: "from at zero",
			req:  SearchRequest{Page: Page{Size: 10, From: 0}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
		})
	}
}

func TestSearchRequest_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		req  SearchRequest
	}{
		{
			name: "invalid match mode",
			req:  SearchRequest{Match: "exact", Page: Page{Size: 10}},
		},
		{
			name: "invalid sort field",
			req:  SearchRequest{Sort: Sort{Field: "invalid"}, Page: Page{Size: 10}},
		},
		{
			name: "invalid sort order",
			req:  SearchRequest{Sort: Sort{Order: "random"}, Page: Page{Size: 10}},
		},
		{
			name: "page size zero",
			req:  SearchRequest{Page: Page{Size: 0}},
		},
		{
			name: "page size negative",
			req:  SearchRequest{Page: Page{Size: -1}},
		},
		{
			name: "page size exceeds max",
			req:  SearchRequest{Page: Page{Size: 101}},
		},
		{
			name: "page from negative",
			req:  SearchRequest{Page: Page{Size: 10, From: -1}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.req.Validate(); err == nil {
				t.Error("expected validation error")
			}
		})
	}
}

func TestSearchRequest_Clone_DeepCopy(t *testing.T) {
	now := time.Now()
	later := now.Add(24 * time.Hour)

	orig := SearchRequest{
		Query:    "test",
		Keywords: "+foo",
		Match:    "all",
		Sort:     Sort{Field: "created_at", Order: "desc"},
		Page:     Page{Size: 20, From: 5},
		Dates: map[string]DateRange{
			"created_at": {After: &now, Before: &later},
		},
		Facets: map[string][]string{
			"mime_type": {"application/pdf", "text/plain"},
		},
	}

	cloned := orig.Clone()

	// Scalar fields match.
	if cloned.Query != orig.Query {
		t.Errorf("query mismatch: %q vs %q", cloned.Query, orig.Query)
	}
	if cloned.Match != orig.Match {
		t.Errorf("match mismatch: %q vs %q", cloned.Match, orig.Match)
	}
	if cloned.Page.Size != orig.Page.Size {
		t.Errorf("page size mismatch: %d vs %d", cloned.Page.Size, orig.Page.Size)
	}
}

func TestSearchRequest_Clone_DatesIndependence(t *testing.T) {
	now := time.Now()
	orig := SearchRequest{
		Dates: map[string]DateRange{
			"created_at": {After: &now},
		},
		Page: Page{Size: 10},
	}

	cloned := orig.Clone()

	// Mutating clone's date pointer should not affect original.
	newTime := now.Add(time.Hour)
	*cloned.Dates["created_at"].After = newTime

	if orig.Dates["created_at"].After.Equal(newTime) {
		t.Error("mutating cloned date affected original — pointers not deep copied")
	}
}

func TestSearchRequest_Clone_FacetsIndependence(t *testing.T) {
	orig := SearchRequest{
		Facets: map[string][]string{
			"mime_type": {"application/pdf"},
		},
		Page: Page{Size: 10},
	}

	cloned := orig.Clone()

	// Mutating clone's facet slice should not affect original.
	cloned.Facets["mime_type"][0] = "text/plain"

	if orig.Facets["mime_type"][0] == "text/plain" {
		t.Error("mutating cloned facet affected original — slices not deep copied")
	}
}

func TestSearchRequest_Clone_NilMaps(t *testing.T) {
	orig := SearchRequest{Page: Page{Size: 10}}

	cloned := orig.Clone()

	if cloned.Dates != nil {
		t.Error("expected nil Dates map in clone")
	}
	if cloned.Facets != nil {
		t.Error("expected nil Facets map in clone")
	}
}

func TestSearchHit_Clone(t *testing.T) {
	orig := SearchHit{VersionID: "v1", DocumentID: "d1", Score: 0.95}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestSearchResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := SearchResponse{
		Hits: []SearchHit{
			{VersionID: "v1", DocumentID: "d1", Summary: "test", Score: 0.95, CreatedAt: time.Now()},
		},
		Total: 1,
		Page:  Page{Size: 20, HasMore: false},
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Hits[0].VersionID != "v1" {
		t.Errorf("hit mutated: %+v", resp.Hits[0])
	}
}

func TestSearchResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := SearchResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestSearchResponse_Clone_HitsIndependence(t *testing.T) {
	orig := SearchResponse{
		Hits: []SearchHit{{VersionID: "v1"}},
	}
	cloned := orig.Clone()

	cloned.Hits[0].VersionID = "mutated"
	if orig.Hits[0].VersionID == "mutated" {
		t.Error("mutating clone hits affected original")
	}
}

func TestSearchResponse_Clone_FacetsIndependence(t *testing.T) {
	orig := SearchResponse{
		Facets: SearchFacets{
			"mime_type": {{Value: "application/pdf", Count: 5}},
		},
	}
	cloned := orig.Clone()

	cloned.Facets["mime_type"][0].Count = 99
	if orig.Facets["mime_type"][0].Count == 99 {
		t.Error("mutating clone facets affected original")
	}
}

func TestSearchResponse_Clone_NilFields(t *testing.T) {
	orig := SearchResponse{Total: 0}
	cloned := orig.Clone()
	if cloned.Hits != nil {
		t.Error("expected nil Hits")
	}
	if cloned.Facets != nil {
		t.Error("expected nil Facets")
	}
}
