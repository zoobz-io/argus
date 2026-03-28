//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	"github.com/zoobz-io/vex"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestSearchDocuments_Success(t *testing.T) {
	searchMock := &argustest.MockDocumentVersionSearch{
		OnSearch: func(_ context.Context, params *models.SearchParams) (*models.SearchResult, error) {
			if params.TenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %q", params.TenantID)
			}
			return &models.SearchResult{
				Hits:  []models.SearchHit{{VersionID: "v1", DocumentID: "d1", Score: 0.95}},
				Total: 1,
			}, nil
		},
	}
	embedMock := &argustest.MockQueryEmbedder{
		OnEmbedQuery: func(_ context.Context, text string) (vex.Vector, error) {
			return vex.Vector{0.1, 0.2, 0.3}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{
		Query: "revenue growth",
		Page:  wire.Page{Size: 20},
	}
	capture := rtesting.ServeRequest(engine, "POST", "/search", body)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.SearchResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 1 || len(resp.Hits) != 1 {
		t.Errorf("unexpected: total=%d hits=%d", resp.Total, len(resp.Hits))
	}
}

func TestSearchDocuments_WithDateRanges(t *testing.T) {
	searchMock := &argustest.MockDocumentVersionSearch{
		OnSearch: func(_ context.Context, params *models.SearchParams) (*models.SearchResult, error) {
			if params.Dates == nil {
				t.Error("expected date ranges to be passed")
			}
			return &models.SearchResult{Total: 0}, nil
		},
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	now := time.Now()
	body := wire.SearchRequest{
		Dates: map[string]wire.DateRange{"created_at": {After: &now}},
		Match: "any",
		Page:  wire.Page{Size: 10},
	}
	capture := rtesting.ServeRequest(engine, "POST", "/search", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestSearchDocuments_EmbedError(t *testing.T) {
	embedMock := &argustest.MockQueryEmbedder{
		OnEmbedQuery: func(_ context.Context, _ string) (vex.Vector, error) { return nil, fmt.Errorf("embed failed") },
	}
	searchMock := &argustest.MockDocumentVersionSearch{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Query: "test", Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/search", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestSearchDocuments_SearchError(t *testing.T) {
	searchMock := &argustest.MockDocumentVersionSearch{
		OnSearch: func(_ context.Context, _ *models.SearchParams) (*models.SearchResult, error) {
			return nil, fmt.Errorf("search failed")
		},
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Keywords: "test", Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/search", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestSearchDocuments_MatchAll(t *testing.T) {
	searchMock := &argustest.MockDocumentVersionSearch{
		OnSearch: func(_ context.Context, params *models.SearchParams) (*models.SearchResult, error) {
			if params.Match != models.MatchAll {
				t.Errorf("expected MatchAll, got %v", params.Match)
			}
			return &models.SearchResult{Total: 0}, nil
		},
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Match: "all", Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/search", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestSearchDocuments_NoQuery(t *testing.T) {
	searchMock := &argustest.MockDocumentVersionSearch{
		OnSearch: func(_ context.Context, params *models.SearchParams) (*models.SearchResult, error) {
			if len(params.QueryVectors) != 0 {
				t.Error("expected no vectors when query is empty")
			}
			return &models.SearchResult{Total: 0}, nil
		},
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Keywords: "+quarterly", Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/search", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestRelatedDocuments_DocNotFound(t *testing.T) {
	docMock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) { return nil, fmt.Errorf("not found") },
	}
	searchMock := &argustest.MockDocumentVersionSearch{}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(docMock), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/documents/d1/related", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestRelatedDocuments_EmbeddingError(t *testing.T) {
	docMock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) { return &models.Document{ID: id}, nil },
	}
	searchMock := &argustest.MockDocumentVersionSearch{
		OnGetDocumentEmbedding: func(_ context.Context, _ string) ([]float32, error) { return nil, fmt.Errorf("embed failed") },
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(docMock), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/documents/d1/related", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestRelatedDocuments_QueryEmbedError(t *testing.T) {
	docMock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) { return &models.Document{ID: id}, nil },
	}
	searchMock := &argustest.MockDocumentVersionSearch{
		OnGetDocumentEmbedding: func(_ context.Context, _ string) ([]float32, error) { return []float32{0.1}, nil },
	}
	embedMock := &argustest.MockQueryEmbedder{
		OnEmbedQuery: func(_ context.Context, _ string) (vex.Vector, error) { return nil, fmt.Errorf("embed failed") },
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(docMock), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Query: "similar to", Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/documents/d1/related", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestRelatedDocuments_SearchError(t *testing.T) {
	docMock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) { return &models.Document{ID: id}, nil },
	}
	searchMock := &argustest.MockDocumentVersionSearch{
		OnGetDocumentEmbedding: func(_ context.Context, _ string) ([]float32, error) { return []float32{0.1}, nil },
		OnSearch:               func(_ context.Context, _ *models.SearchParams) (*models.SearchResult, error) { return nil, fmt.Errorf("search failed") },
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(docMock), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/documents/d1/related", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestRelatedDocuments_WithQueryVector(t *testing.T) {
	docMock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) { return &models.Document{ID: id}, nil },
	}
	searchMock := &argustest.MockDocumentVersionSearch{
		OnGetDocumentEmbedding: func(_ context.Context, _ string) ([]float32, error) { return []float32{0.1}, nil },
		OnSearch: func(_ context.Context, params *models.SearchParams) (*models.SearchResult, error) {
			if len(params.QueryVectors) != 2 {
				t.Errorf("expected 2 vectors (doc + query), got %d", len(params.QueryVectors))
			}
			return &models.SearchResult{Total: 0}, nil
		},
	}
	embedMock := &argustest.MockQueryEmbedder{
		OnEmbedQuery: func(_ context.Context, _ string) (vex.Vector, error) { return vex.Vector{0.2, 0.3}, nil },
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(docMock), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Query: "revenue", Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/documents/d1/related", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestRelatedDocuments_Success(t *testing.T) {
	docMock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) {
			return &models.Document{ID: id}, nil
		},
	}
	searchMock := &argustest.MockDocumentVersionSearch{
		OnGetDocumentEmbedding: func(_ context.Context, _ string) ([]float32, error) {
			return []float32{0.5, 0.6}, nil
		},
		OnSearch: func(_ context.Context, params *models.SearchParams) (*models.SearchResult, error) {
			if len(params.QueryVectors) == 0 {
				t.Error("expected at least the document embedding vector")
			}
			return &models.SearchResult{Total: 2, Hits: []models.SearchHit{{VersionID: "v2"}, {VersionID: "v3"}}}, nil
		},
	}
	embedMock := &argustest.MockQueryEmbedder{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(docMock), argustest.WithAPIDocumentVersionSearch(searchMock), argustest.WithAPIQueryEmbedder(embedMock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.SearchRequest{Page: wire.Page{Size: 10}}
	capture := rtesting.ServeRequest(engine, "POST", "/documents/d1/related", body)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.SearchResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Total != 2 {
		t.Errorf("expected 2 related, got %d", resp.Total)
	}
}
