package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// TopicToResponse transforms a Topic model to an API response.
func TopicToResponse(t *models.Topic) wire.TopicResponse {
	return wire.TopicResponse{
		ID:          t.ID,
		Name:        t.Name,
		Description: t.Description,
	}
}

// TopicsToListResponse transforms a slice of Topic models to a list response.
func TopicsToListResponse(topics []*models.Topic) wire.TopicListResponse {
	items := make([]wire.TopicResponse, len(topics))
	for i, t := range topics {
		items[i] = TopicToResponse(t)
	}
	return wire.TopicListResponse{
		Topics: items,
	}
}
