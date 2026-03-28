package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// SearchResultToResponse transforms a domain search result to an API response.
func SearchResultToResponse(result *models.SearchResult, page wire.Page) wire.SearchResponse {
	hits := make([]wire.SearchHit, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = wire.SearchHit{
			CreatedAt:    hit.CreatedAt,
			UpdatedAt:    hit.UpdatedAt,
			VersionID:    hit.VersionID,
			DocumentID:   hit.DocumentID,
			DocumentName: hit.DocumentName,
			MimeType:     hit.MimeType,
			Summary:      hit.Summary,
			Score:        hit.Score,
		}
	}

	facets := make(wire.SearchFacets, len(result.Facets))
	for name, buckets := range result.Facets {
		counts := make([]wire.FacetCount, len(buckets))
		for i, f := range buckets {
			counts[i] = wire.FacetCount{Value: f.Value, Count: f.Count}
		}
		facets[name] = counts
	}

	return wire.SearchResponse{
		Hits:  hits,
		Total: result.Total,
		Page: wire.Page{
			Size:    page.Size,
			From:    page.From,
			HasMore: int64(page.From+page.Size) < result.Total,
		},
		Facets: facets,
	}
}
