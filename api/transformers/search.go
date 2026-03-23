package transformers

import (
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// SearchResultToResponse transforms an OpenSearch result to an API response.
func SearchResultToResponse(result *grub.SearchResult[models.DocumentVersionIndex]) wire.SearchResponse {
	hits := make([]wire.SearchHit, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = wire.SearchHit{
			VersionID:    hit.Content.VersionID,
			DocumentID:   hit.Content.DocumentID,
			DocumentName: hit.Content.DocumentName,
			MimeType:     hit.Content.MimeType,
			Summary:      hit.Content.Summary,
			Score:        hit.Score,
		}
	}
	return wire.SearchResponse{
		Hits:  hits,
		Total: result.Total,
	}
}
