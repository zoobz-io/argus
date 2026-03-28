package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// TagToAdminResponse converts a tag model to an admin API response.
func TagToAdminResponse(t *models.Tag) wire.AdminTagResponse {
	return wire.AdminTagResponse{
		ID:          t.ID,
		TenantID:    t.TenantID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// TagsToAdminList converts a slice of tags to an admin list response.
func TagsToAdminList(tags []*models.Tag) wire.AdminTagListResponse {
	items := make([]wire.AdminTagResponse, len(tags))
	for i, t := range tags {
		items[i] = TagToAdminResponse(t)
	}
	return wire.AdminTagListResponse{
		Tags: items,
	}
}
