package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// TagToResponse transforms a Tag model to an API response.
func TagToResponse(t *models.Tag) wire.TagResponse {
	return wire.TagResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
	}
}

// TagsToListResponse transforms a slice of Tag models to a list response.
func TagsToListResponse(tags []*models.Tag) wire.TagListResponse {
	items := make([]wire.TagResponse, len(tags))
	for i, t := range tags {
		items[i] = TagToResponse(t)
	}
	return wire.TagListResponse{
		Tags: items,
	}
}
