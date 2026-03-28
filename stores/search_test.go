//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

func newTestSearch(t *testing.T, mock *mockSearchProvider) *DocumentVersionSearch {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewDocumentVersionSearch(mock)
}

func testVersionIndex() models.DocumentVersionIndex {
	return models.DocumentVersionIndex{
		VersionID:    "ver-1",
		DocumentID:   "doc-1",
		TenantID:     "t-1",
		ProviderID:   "p-1",
		DocumentName: "report.pdf",
		MimeType:     "application/pdf",
		Content:      "test content",
		Summary:      "a summary",
		Language:     "en",
		CreatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		Embedding:    []float32{0.1, 0.2},
	}
}

func TestDocumentVersionSearch_IndexVersion(t *testing.T) {
	var indexedID string
	mock := &mockSearchProvider{
		OnIndex: func(_ context.Context, _, id string, _ []byte) error {
			indexedID = id
			return nil
		},
	}
	store := newTestSearch(t, mock)
	idx := testVersionIndex()

	err := store.IndexVersion(context.Background(), &idx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if indexedID != "doc-1" {
		t.Errorf("indexed by DocumentID: got %q, want %q", indexedID, "doc-1")
	}
}

func TestDocumentVersionSearch_IndexVersion_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnIndex: func(_ context.Context, _, _ string, _ []byte) error {
			return errors.New("index failure")
		},
	}
	store := newTestSearch(t, mock)
	idx := testVersionIndex()

	err := store.IndexVersion(context.Background(), &idx)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentVersionSearch_DeleteDocument(t *testing.T) {
	var deletedID string
	mock := &mockSearchProvider{
		OnDelete: func(_ context.Context, _, id string) error {
			deletedID = id
			return nil
		},
	}
	store := newTestSearch(t, mock)

	err := store.DeleteDocument(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if deletedID != "doc-1" {
		t.Errorf("deleted ID: got %q, want %q", deletedID, "doc-1")
	}
}

func TestDocumentVersionSearch_GetDocumentEmbedding(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnGet: func(_ context.Context, _, _ string) ([]byte, error) {
			return jsonBytes(t, idx), nil
		},
	}
	store := newTestSearch(t, mock)

	embedding, err := store.GetDocumentEmbedding(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(embedding) != 2 {
		t.Errorf("embedding len: got %d, want 2", len(embedding))
	}
}

func TestDocumentVersionSearch_GetDocumentEmbedding_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnGet: func(_ context.Context, _, _ string) ([]byte, error) {
			return nil, errors.New("not found")
		},
	}
	store := newTestSearch(t, mock)

	_, err := store.GetDocumentEmbedding(context.Background(), "doc-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentVersionSearch_Search_Keyword(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 1.5}},
				Total: 1,
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID: "t-1",
		Query:    "test",
		Size:     10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
	if len(result.Hits) != 1 {
		t.Fatalf("Hits: got %d, want 1", len(result.Hits))
	}
	if result.Hits[0].DocumentID != "doc-1" {
		t.Errorf("DocumentID: got %q, want %q", result.Hits[0].DocumentID, "doc-1")
	}
}

func TestDocumentVersionSearch_Search_Semantic(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 0.9}},
				Total: 1,
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID:     "t-1",
		QueryVectors: [][]float32{{0.1, 0.2}},
		Size:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
}

func TestDocumentVersionSearch_Search_Hybrid(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 2.0}},
				Total: 1,
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID:     "t-1",
		Query:        "test",
		QueryVectors: [][]float32{{0.1, 0.2}},
		Size:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
}

func TestDocumentVersionSearch_Search_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return nil, errors.New("search failed")
		},
	}
	store := newTestSearch(t, mock)

	_, err := store.Search(context.Background(), &models.SearchParams{
		TenantID: "t-1",
		Query:    "test",
		Size:     10,
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentVersionSearch_Search_WithAggregations(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 1.0}},
				Total: 1,
				Aggregations: map[string]any{
					"mime_types": map[string]any{
						"buckets": []any{
							map[string]any{"key": "application/pdf", "doc_count": float64(5)},
						},
					},
					"topics": map[string]any{
						"buckets": []any{
							map[string]any{"key": "Finance", "doc_count": float64(3)},
						},
					},
				},
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID: "t-1",
		Query:    "test",
		Size:     10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Facets) != 2 {
		t.Errorf("Facets: got %d groups, want 2", len(result.Facets))
	}
	if len(result.Facets["mime_types"]) != 1 {
		t.Errorf("mime_types facets: got %d, want 1", len(result.Facets["mime_types"]))
	}
}

func TestDocumentVersionSearch_Search_SemanticMultiVector(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 0.9}},
				Total: 1,
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	// Multiple vectors triggers the multi-KNN path (Should by default).
	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID:     "t-1",
		QueryVectors: [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		Size:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
}

func TestDocumentVersionSearch_Search_SemanticMatchAll(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 0.8}},
				Total: 1,
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	// Multiple vectors + MatchAll triggers Bool.Must path.
	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID:     "t-1",
		QueryVectors: [][]float32{{0.1, 0.2}, {0.3, 0.4}},
		Match:        models.MatchAll,
		Size:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
}

func TestDocumentVersionSearch_Search_HybridMatchAll(t *testing.T) {
	idx := testVersionIndex()
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Hits:  []grub.SearchHit{{ID: "doc-1", Source: jsonBytes(t, idx), Score: 2.5}},
				Total: 1,
			}, nil
		},
	}
	store := newTestSearch(t, mock)

	// Keyword + vector + MatchAll triggers Bool.Must(keyword, semantic).
	result, err := store.Search(context.Background(), &models.SearchParams{
		TenantID:     "t-1",
		Query:        "test",
		QueryVectors: [][]float32{{0.1, 0.2}},
		Match:        models.MatchAll,
		Size:         10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("Total: got %d, want 1", result.Total)
	}
}
