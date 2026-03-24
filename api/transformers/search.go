package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// SearchResultToResponse transforms a domain search result to an API response.
func SearchResultToResponse(result *models.SearchResult) wire.SearchResponse {
	hits := make([]wire.SearchHit, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = wire.SearchHit{
			VersionID:    hit.VersionID,
			DocumentID:   hit.DocumentID,
			DocumentName: hit.DocumentName,
			MimeType:     hit.MimeType,
			Summary:      hit.Summary,
			Score:        hit.Score,
		}
	}
	return wire.SearchResponse{
		Hits:  hits,
		Total: result.Total,
	}
}
