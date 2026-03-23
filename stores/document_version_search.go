package stores

import (
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/argus/models"
)

// DocumentVersionSearch wraps a search index for document versions.
type DocumentVersionSearch struct {
	*sum.Search[models.DocumentVersionIndex]
}

// NewDocumentVersionSearch creates a new document version search store.
func NewDocumentVersionSearch(provider grub.SearchProvider) (*DocumentVersionSearch, error) {
	search, err := sum.NewSearch[models.DocumentVersionIndex](provider, "document_versions")
	if err != nil {
		return nil, err
	}
	return &DocumentVersionSearch{Search: search}, nil
}
