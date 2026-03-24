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
func NewDocumentVersionSearch(provider grub.SearchProvider) (*DocumentVersionSearch, error) {
	search, err := sum.NewSearch[models.DocumentVersionIndex](provider, "document_versions")
	if err != nil {
		return nil, err
	}
	qb, err := lucene.New[models.DocumentVersionIndex]()
	if err != nil {
		return nil, fmt.Errorf("creating search builder: %w", err)
	}
	return &DocumentVersionSearch{index: search, qb: qb}, nil
}

// Search performs a full-text search across document version content.
func (s *DocumentVersionSearch) Search(ctx context.Context, query string, size int, from int) (*models.SearchResult, error) {
	search := lucene.NewSearch().
		Query(s.qb.Match("content", query)).
		Size(size).
		From(from)
	result, err := s.index.Execute(ctx, search)
	if err != nil {
		return nil, fmt.Errorf("executing search: %w", err)
	}
	hits := make([]models.SearchHit, len(result.Hits))
	for i, hit := range result.Hits {
		hits[i] = models.SearchHit{
			VersionID:    hit.Content.VersionID,
			DocumentID:   hit.Content.DocumentID,
			DocumentName: hit.Content.DocumentName,
			MimeType:     hit.Content.MimeType,
			Summary:      hit.Content.Summary,
			Score:        hit.Score,
		}
	}
	return &models.SearchResult{
		Hits:  hits,
		Total: result.Total,
	}, nil
}

// IndexVersion indexes a document version in the search index.
func (s *DocumentVersionSearch) IndexVersion(ctx context.Context, version *models.DocumentVersionIndex) error {
	id := fmt.Sprintf("%d", version.VersionID)
	return s.index.Index(ctx, id, version)
}

// DeleteVersion removes a document version from the search index.
func (s *DocumentVersionSearch) DeleteVersion(ctx context.Context, versionID int64) error {
	id := fmt.Sprintf("%d", versionID)
	return s.index.Delete(ctx, id)
}
