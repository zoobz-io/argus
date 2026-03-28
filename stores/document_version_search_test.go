package stores

import (
	"testing"
	"time"

	"github.com/zoobz-io/lucene"

	"github.com/zoobz-io/argus/models"
)

func newTestSearchStore(t *testing.T) *DocumentVersionSearch {
	t.Helper()
	return &DocumentVersionSearch{qb: lucene.New[models.DocumentVersionIndex]()}
}

// --- extractBuckets tests (package-level function, pure logic) ---

func TestExtractBuckets_NilAggs(t *testing.T) {
	result := extractBuckets(nil, "topics")
	if result != nil {
		t.Errorf("expected nil, got %v", result)
	}
}

func TestExtractBuckets_MissingName(t *testing.T) {
	aggs := map[string]any{}
	result := extractBuckets(aggs, "topics")
	if result != nil {
		t.Errorf("expected nil for missing name, got %v", result)
	}
}

func TestExtractBuckets_WrongAggType(t *testing.T) {
	aggs := map[string]any{"topics": "not-a-map"}
	result := extractBuckets(aggs, "topics")
	if result != nil {
		t.Errorf("expected nil for wrong agg type, got %v", result)
	}
}

func TestExtractBuckets_MissingBucketsKey(t *testing.T) {
	aggs := map[string]any{"topics": map[string]any{"no_buckets": true}}
	result := extractBuckets(aggs, "topics")
	if result != nil {
		t.Errorf("expected nil for missing buckets key, got %v", result)
	}
}

func TestExtractBuckets_WrongBucketsType(t *testing.T) {
	aggs := map[string]any{"topics": map[string]any{"buckets": "not-an-array"}}
	result := extractBuckets(aggs, "topics")
	if result != nil {
		t.Errorf("expected nil for wrong buckets type, got %v", result)
	}
}

func TestExtractBuckets_ValidBuckets(t *testing.T) {
	aggs := map[string]any{
		"topics": map[string]any{
			"buckets": []any{
				map[string]any{"key": "Finance", "doc_count": float64(42)},
				map[string]any{"key": "Legal", "doc_count": float64(17)},
			},
		},
	}

	result := extractBuckets(aggs, "topics")
	if len(result) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(result))
	}
	if result[0].Value != "Finance" || result[0].Count != 42 {
		t.Errorf("bucket[0]: got {%q, %d}, want {Finance, 42}", result[0].Value, result[0].Count)
	}
	if result[1].Value != "Legal" || result[1].Count != 17 {
		t.Errorf("bucket[1]: got {%q, %d}, want {Legal, 17}", result[1].Value, result[1].Count)
	}
}

func TestExtractBuckets_MalformedBucket(t *testing.T) {
	aggs := map[string]any{
		"topics": map[string]any{
			"buckets": []any{
				"not-a-map",
				map[string]any{"key": "Valid", "doc_count": float64(5)},
			},
		},
	}

	result := extractBuckets(aggs, "topics")
	if len(result) != 1 {
		t.Fatalf("expected 1 valid bucket, got %d", len(result))
	}
	if result[0].Value != "Valid" {
		t.Errorf("bucket value: got %q, want %q", result[0].Value, "Valid")
	}
}

func TestExtractBuckets_MissingDocCount(t *testing.T) {
	aggs := map[string]any{
		"topics": map[string]any{
			"buckets": []any{
				map[string]any{"key": "NoCount"},
			},
		},
	}

	result := extractBuckets(aggs, "topics")
	if len(result) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result))
	}
	if result[0].Count != 0 {
		t.Errorf("expected count 0 for missing doc_count, got %d", result[0].Count)
	}
}

// --- extractFacets tests ---

func TestExtractFacets_AllPresent(t *testing.T) {
	s := newTestSearchStore(t)
	aggs := map[string]any{
		"mime_types": map[string]any{
			"buckets": []any{
				map[string]any{"key": "application/pdf", "doc_count": float64(10)},
			},
		},
		"topics": map[string]any{
			"buckets": []any{
				map[string]any{"key": "Finance", "doc_count": float64(5)},
			},
		},
		"tags": map[string]any{
			"buckets": []any{
				map[string]any{"key": "Important", "doc_count": float64(3)},
			},
		},
		"languages": map[string]any{
			"buckets": []any{
				map[string]any{"key": "en", "doc_count": float64(20)},
			},
		},
	}

	facets := s.extractFacets(aggs)
	if len(facets) != 4 {
		t.Errorf("expected 4 facet groups, got %d", len(facets))
	}
	if len(facets["topics"]) != 1 || facets["topics"][0].Value != "Finance" {
		t.Errorf("topics facet: got %v", facets["topics"])
	}
}

func TestExtractFacets_Empty(t *testing.T) {
	s := newTestSearchStore(t)
	facets := s.extractFacets(map[string]any{})
	if len(facets) != 0 {
		t.Errorf("expected 0 facets for empty aggs, got %d", len(facets))
	}
}

// --- buildFilters tests ---

func TestBuildFilters_TenantID(t *testing.T) {
	s := newTestSearchStore(t)
	params := &models.SearchParams{TenantID: "tenant-1"}
	filters := s.buildFilters(params)
	if len(filters) != 1 {
		t.Errorf("expected 1 filter for tenant_id, got %d", len(filters))
	}
}

func TestBuildFilters_NoTenantID(t *testing.T) {
	s := newTestSearchStore(t)
	params := &models.SearchParams{}
	filters := s.buildFilters(params)
	if len(filters) != 0 {
		t.Errorf("expected 0 filters, got %d", len(filters))
	}
}

func TestBuildFilters_Facets(t *testing.T) {
	s := newTestSearchStore(t)
	params := &models.SearchParams{
		Facets: map[string][]string{
			"topics": {"Finance", "Legal"},
			"tags":   {"Important"},
		},
	}
	filters := s.buildFilters(params)
	if len(filters) != 2 {
		t.Errorf("expected 2 facet filters, got %d", len(filters))
	}
}

func TestBuildFilters_InvalidFacet(t *testing.T) {
	s := newTestSearchStore(t)
	params := &models.SearchParams{
		Facets: map[string][]string{
			"unknown_facet": {"value"},
		},
	}
	filters := s.buildFilters(params)
	if len(filters) != 0 {
		t.Errorf("expected 0 filters for unknown facet, got %d", len(filters))
	}
}

func TestBuildFilters_EmptyFacetValues(t *testing.T) {
	s := newTestSearchStore(t)
	params := &models.SearchParams{
		Facets: map[string][]string{
			"topics": {},
		},
	}
	filters := s.buildFilters(params)
	if len(filters) != 0 {
		t.Errorf("expected 0 filters for empty facet values, got %d", len(filters))
	}
}

func TestBuildFilters_DateRanges(t *testing.T) {
	s := newTestSearchStore(t)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	before := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	params := &models.SearchParams{
		Dates: map[string]models.DateRange{
			"created_at": {After: &after, Before: &before},
		},
	}
	filters := s.buildFilters(params)
	if len(filters) != 2 {
		t.Errorf("expected 2 range filters (gte + lte), got %d", len(filters))
	}
}

func TestBuildFilters_DateRangeAfterOnly(t *testing.T) {
	s := newTestSearchStore(t)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	params := &models.SearchParams{
		Dates: map[string]models.DateRange{
			"created_at": {After: &after},
		},
	}
	filters := s.buildFilters(params)
	if len(filters) != 1 {
		t.Errorf("expected 1 range filter (gte only), got %d", len(filters))
	}
}

func TestBuildFilters_InvalidDateField(t *testing.T) {
	s := newTestSearchStore(t)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	params := &models.SearchParams{
		Dates: map[string]models.DateRange{
			"unknown_date": {After: &after},
		},
	}
	filters := s.buildFilters(params)
	if len(filters) != 0 {
		t.Errorf("expected 0 filters for unknown date field, got %d", len(filters))
	}
}

func TestBuildFilters_Combined(t *testing.T) {
	s := newTestSearchStore(t)
	after := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	params := &models.SearchParams{
		TenantID: "tenant-1",
		Facets:   map[string][]string{"topics": {"Finance"}},
		Dates:    map[string]models.DateRange{"created_at": {After: &after}},
	}
	filters := s.buildFilters(params)
	if len(filters) != 3 {
		t.Errorf("expected 3 filters (tenant + facet + date), got %d", len(filters))
	}
}

// --- buildAggregations tests ---

func TestBuildAggregations(t *testing.T) {
	s := newTestSearchStore(t)
	aggs := s.buildAggregations()
	if len(aggs) != 4 {
		t.Errorf("expected 4 aggregations, got %d", len(aggs))
	}
}

// --- applySort tests ---

func TestApplySort_ValidField(t *testing.T) {
	s := newTestSearchStore(t)
	search := lucene.NewSearch().Size(10)

	result := s.applySort(search, models.SearchSort{Field: "created_at", Order: "asc"})
	if result == nil {
		t.Fatal("expected non-nil search")
	}
}

func TestApplySort_UnknownField(t *testing.T) {
	s := newTestSearchStore(t)
	search := lucene.NewSearch().Size(10)

	result := s.applySort(search, models.SearchSort{Field: "unknown_field"})
	// Should return the search unchanged.
	if result != search {
		t.Error("expected same search object for unknown field")
	}
}

func TestApplySort_DefaultOrder(t *testing.T) {
	s := newTestSearchStore(t)
	search := lucene.NewSearch().Size(10)

	// Empty order should default to "desc".
	result := s.applySort(search, models.SearchSort{Field: "created_at", Order: ""})
	if result == nil {
		t.Fatal("expected non-nil search")
	}
}

func TestApplySort_NameField(t *testing.T) {
	s := newTestSearchStore(t)
	search := lucene.NewSearch().Size(10)

	// "name" maps to "document_name.keyword".
	result := s.applySort(search, models.SearchSort{Field: "name", Order: "asc"})
	if result == nil {
		t.Fatal("expected non-nil search")
	}
}
