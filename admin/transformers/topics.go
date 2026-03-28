package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// TopicToAdminResponse converts a topic model to an admin API response.
func TopicToAdminResponse(t *models.Topic) wire.AdminTopicResponse {
	return wire.AdminTopicResponse{
		ID:          t.ID,
		TenantID:    t.TenantID,
		Name:        t.Name,
		Description: t.Description,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// TopicsToAdminList converts a slice of topics to an admin list response.
func TopicsToAdminList(topics []*models.Topic) wire.AdminTopicListResponse {
	items := make([]wire.AdminTopicResponse, len(topics))
	for i, t := range topics {
		items[i] = TopicToAdminResponse(t)
	}
	return wire.AdminTopicListResponse{
		Topics: items,
	}
}
