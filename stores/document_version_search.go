package stores

import (
	"context"
	"fmt"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch wraps a search index for document versions.
type DocumentVersionSearch struct {
	index *sum.Search[models.DocumentVersionIndex]
	qb    *lucene.Builder[models.DocumentVersionIndex]
}

// NewDocumentVersionSearch creates a new document version search store.
func NewDocumentVersionSearch(provider grub.SearchProvider) *DocumentVersionSearch {
	return &DocumentVersionSearch{
		index: sum.NewSearch[models.DocumentVersionIndex](provider, "document_versions"),
		qb:    lucene.New[models.DocumentVersionIndex](),
	}
}

// Search performs a keyword, semantic, or hybrid search based on the provided params.
func (s *DocumentVersionSearch) Search(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error) {
	hasKeyword := params.Query != ""
	hasSemantic := len(params.QueryVectors) > 0

	switch {
	case hasKeyword && hasSemantic:
		return s.hybridSearch(ctx, params)
	case hasSemantic:
		return s.semanticSearch(ctx, params)
	default:
		return s.keywordSearch(ctx, params)
	}
}

// GetDocumentEmbedding retrieves the embedding vector for a document.
func (s *DocumentVersionSearch) GetDocumentEmbedding(ctx context.Context, documentID string) ([]float32, error) {
	doc, err := s.index.Get(ctx, documentID)
	if err != nil {
		return nil, fmt.Errorf("fetching document embedding: %w", err)
	}
	return doc.Content.Embedding, nil
}

// IndexVersion indexes a document version in the search index.
func (s *DocumentVersionSearch) IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error {
	return s.index.Index(ctx, version.DocumentID, version)
}

// DeleteDocument removes a document from the search index.
func (s *DocumentVersionSearch) DeleteDocument(ctx context.Context, documentID string) error {
	return s.index.Delete(ctx, documentID)
}

func (s *DocumentVersionSearch) keywordSearch(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error) {
	boolQ := s.qb.Bool().Filter(s.buildFilters(params)...)
	if params.Query != "" {
		boolQ = boolQ.Must(s.qb.MultiMatch(params.Query, "content", "summary"))
	}

	search := lucene.NewSearch().
		Query(boolQ).
		Aggs(s.buildAggregations()...).
		Size(params.Size).
		From(params.From)

	search = s.applySort(search, params.Sort)

	return s.executeSearch(ctx, search)
}

func (s *DocumentVersionSearch) semanticSearch(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error) {
	filter := s.qb.Bool().Filter(s.buildFilters(params)...)
	q := s.buildSemanticQuery(params, filter)

	search := lucene.NewSearch().
		Query(q).
		Aggs(s.buildAggregations()...).
		Size(params.Size).
		From(params.From)

	return s.executeSearch(ctx, search)
}

func (s *DocumentVersionSearch) buildSemanticQuery(params *models.SearchParams, filter *lucene.BoolQuery) lucene.Query {
	knns := make([]lucene.Query, len(params.QueryVectors))
	for i, vec := range params.QueryVectors {
		knns[i] = s.qb.Knn("embedding", vec).
			K(params.Size + params.From).
			Filter(filter)
	}
	if len(knns) == 1 {
		return knns[0]
	}
	if params.Match == models.MatchAll {
		return s.qb.Bool().Must(knns...)
	}
	return s.qb.Bool().Should(knns...)
}

func (s *DocumentVersionSearch) hybridSearch(ctx context.Context, params *models.SearchParams) (*models.SearchResult, error) {
	filter := s.qb.Bool().Filter(s.buildFilters(params)...)
	semanticQ := s.buildSemanticQuery(params, filter)

	keywordQ := s.qb.Bool().
		Must(s.qb.MultiMatch(params.Query, "content", "summary")).
		Filter(s.buildFilters(params)...)

	var q lucene.Query
	if params.Match == models.MatchAll {
		q = s.qb.Bool().Must(keywordQ, semanticQ)
	} else {
		q = s.qb.Hybrid(keywordQ, semanticQ)
	}

	search := lucene.NewSearch().
		Query(q).
		Aggs(s.buildAggregations()...).
		Size(params.Size).
		From(params.From)

	return s.executeSearch(ctx, search)
}

// facetFields maps request facet names to their OpenSearch field names.
var facetFields = map[string]string{
	"mime_types": "mime_type",
	"topics":    "topics",
	"tags":      "tags",
	"languages": "language",
}

// dateFields maps request date names to their OpenSearch field names.
var dateFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
}

func (s *DocumentVersionSearch) buildFilters(params *models.SearchParams) []lucene.Query {
	filters := make([]lucene.Query, 0, 1+len(params.Facets)+len(params.Dates))

	if params.TenantID != "" {
		filters = append(filters, s.qb.Term("tenant_id", params.TenantID))
	}
	for facet, values := range params.Facets {
		field, ok := facetFields[facet]
		if !ok || len(values) == 0 {
			continue
		}
		vals := make([]any, len(values))
		for i, v := range values {
			vals[i] = v
		}
		filters = append(filters, s.qb.Terms(field, vals...))
	}
	for name, dr := range params.Dates {
		field, ok := dateFields[name]
		if !ok {
			continue
		}
		if dr.After != nil {
			filters = append(filters, s.qb.Range(field).Gte(dr.After.Format("2006-01-02T15:04:05Z")))
		}
		if dr.Before != nil {
			filters = append(filters, s.qb.Range(field).Lte(dr.Before.Format("2006-01-02T15:04:05Z")))
		}
	}

	return filters
}

func (s *DocumentVersionSearch) buildAggregations() []lucene.Aggregation {
	return []lucene.Aggregation{
		s.qb.TermsAgg("mime_types", "mime_type").Size(50),
		s.qb.TermsAgg("topics", "topics").Size(100),
		s.qb.TermsAgg("tags", "tags").Size(100),
		s.qb.TermsAgg("languages", "language").Size(50),
	}
}

// sortFields maps request sort field names to their OpenSearch field names.
var sortFields = map[string]string{
	"created_at": "created_at",
	"updated_at": "updated_at",
	"name":       "document_name.keyword",
}

func (s *DocumentVersionSearch) applySort(search *lucene.Search, sort models.SearchSort) *lucene.Search {
	field, ok := sortFields[sort.Field]
	if !ok {
		return search
	}
	order := sort.Order
	if order == "" {
		order = "desc"
	}
	return search.Sort(lucene.SortField{Field: field, Order: order})
}

func (s *DocumentVersionSearch) executeSearch(ctx context.Context, search *lucene.Search) (*models.SearchResult, error) {
	result, err := s.index.Execute(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("executing search: %w", err)
	}

	hits := make([]models.SearchHit, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = models.SearchHit{
			CreatedAt:    hit.Content.CreatedAt,
			UpdatedAt:    hit.Content.UpdatedAt,
			VersionID:    hit.Content.VersionID,
			DocumentID:   hit.Content.DocumentID,
			DocumentName: hit.Content.DocumentName,
			MimeType:     hit.Content.MimeType,
			Summary:      hit.Content.Summary,
			Score:        hit.Score,
		}
	}
	return &models.SearchResult{
		Hits:   hits,
		Total:  result.Total,
		Facets: s.extractFacets(result.Aggregations),
	}, nil
}

func (s *DocumentVersionSearch) extractFacets(aggs map[string]any) models.SearchFacets {
	facets := make(models.SearchFacets)
	for _, name := range []string{"mime_types", "topics", "tags", "languages"} {
		if buckets := extractBuckets(aggs, name); buckets != nil {
			facets[name] = buckets
		}
	}
	return facets
}

func extractBuckets(aggs map[string]any, name string) []models.FacetCount {
	if aggs == nil {
		return nil
	}
	aggData, ok := aggs[name]
	if !ok {
		return nil
	}
	aggMap, ok := aggData.(map[string]any)
	if !ok {
		return nil
	}
	bucketsRaw, ok := aggMap["buckets"]
	if !ok {
		return nil
	}
	buckets, ok := bucketsRaw.([]any)
	if !ok {
		return nil
	}
	counts := make([]models.FacetCount, 0, len(buckets))
	for _, b := range buckets {
		bucket, ok := b.(map[string]any)
		if !ok {
			continue
		}
		key := fmt.Sprintf("%v", bucket["key"])
		var count int64
		if dc, ok := bucket["doc_count"].(float64); ok {
			count = int64(dc)
		}
		counts = append(counts, models.FacetCount{Value: key, Count: count})
	}
	return counts
}
