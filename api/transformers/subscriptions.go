package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// SubscriptionToResponse transforms a Subscription model to an API response.
func SubscriptionToResponse(s *models.Subscription) wire.SubscriptionResponse {
	return wire.SubscriptionResponse{
		ID:        s.ID,
		UserID:    s.UserID,
		EventType: s.EventType,
		Channel:   s.Channel,
	}
}

// SubscriptionsToResponse transforms a slice of Subscription models to responses.
func SubscriptionsToResponse(subs []*models.Subscription) []wire.SubscriptionResponse {
	result := make([]wire.SubscriptionResponse, len(subs))
	for i, s := range subs {
		result[i] = SubscriptionToResponse(s)
	}
	return result
}
