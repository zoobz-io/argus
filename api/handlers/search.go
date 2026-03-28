package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/transformers"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

func toDateRanges(dates map[string]wire.DateRange) map[string]models.DateRange {
	if len(dates) == 0 {
		return nil
	}
	out := make(map[string]models.DateRange, len(dates))
	for k, v := range dates {
		out[k] = models.DateRange{After: v.After, Before: v.Before}
	}
	return out
}

func toMatchMode(match string) models.MatchMode {
	if match == "all" {
		return models.MatchAll
	}
	return models.MatchAny
}

func embedQuery(r *rocco.Request[wire.SearchRequest]) ([]float32, error) {
	if r.Body.Query == "" {
		return nil, nil
	}
	embedder := sum.MustUse[contracts.QueryEmbedder](r)
	return embedder.EmbedQuery(r, r.Body.Query)
}

var searchDocuments = rocco.POST[wire.SearchRequest, wire.SearchResponse]("/search", func(r *rocco.Request[wire.SearchRequest]) (wire.SearchResponse, error) {
	tid := tenantID(r.Identity)

	vec, err := embedQuery(r)
	if err != nil {
		return wire.SearchResponse{}, err
	}
	var vectors [][]float32
	if vec != nil {
		vectors = append(vectors, vec)
	}

	params := &models.SearchParams{
		TenantID:     tid,
		Query:        r.Body.Keywords,
		QueryVectors: vectors,
		Match:        toMatchMode(r.Body.Match),
		Facets:       r.Body.Facets,
		Dates:        toDateRanges(r.Body.Dates),
		Sort:         models.SearchSort{Field: r.Body.Sort.Field, Order: r.Body.Sort.Order},
		Size:         r.Body.Page.Size,
		From:         r.Body.Page.From,
	}

	store := sum.MustUse[contracts.DocumentVersionSearch](r)
	result, err := store.Search(r, params)
	if err != nil {
		return wire.SearchResponse{}, err
	}
	return transformers.SearchResultToResponse(result, r.Body.Page), nil
}).
	WithSummary("Search documents").
	WithTags("search").
	WithAuthentication().
	WithErrors(rocco.ErrValidationFailed)

var relatedDocuments = rocco.POST[wire.SearchRequest, wire.SearchResponse]("/documents/{id}/related", func(r *rocco.Request[wire.SearchRequest]) (wire.SearchResponse, error) {
	tid := tenantID(r.Identity)
	docID := pathID(r.Params, "id")

	docs := sum.MustUse[contracts.Documents](r)
	doc, err := docs.GetDocument(r, docID)
	if err != nil {
		return wire.SearchResponse{}, ErrDocumentNotFound
	}
	// Start with the document's embedding.
	store := sum.MustUse[contracts.DocumentVersionSearch](r)
	docVector, err := store.GetDocumentEmbedding(r, doc.ID)
	if err != nil {
		return wire.SearchResponse{}, err
	}
	vectors := [][]float32{docVector}

	// Also embed the user's semantic query if provided.
	vec, err := embedQuery(r)
	if err != nil {
		return wire.SearchResponse{}, err
	}
	if vec != nil {
		vectors = append(vectors, vec)
	}

	params := &models.SearchParams{
		TenantID:     tid,
		Query:        r.Body.Keywords,
		QueryVectors: vectors,
		Match:        toMatchMode(r.Body.Match),
		Facets:       r.Body.Facets,
		Dates:        toDateRanges(r.Body.Dates),
		Sort:         models.SearchSort{Field: r.Body.Sort.Field, Order: r.Body.Sort.Order},
		Size:         r.Body.Page.Size,
		From:         r.Body.Page.From,
	}

	result, err := store.Search(r, params)
	if err != nil {
		return wire.SearchResponse{}, err
	}
	return transformers.SearchResultToResponse(result, r.Body.Page), nil
}).
	WithSummary("Find related documents").
	WithTags("search").
	WithPathParams("id").
	WithAuthentication().
	WithErrors(ErrDocumentNotFound, rocco.ErrValidationFailed)
