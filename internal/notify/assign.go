package notify

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/pipz"

	"github.com/zoobz-io/argus/models"
)

func newAssignStage() pipz.Chainable[*FanOutItem] {
	return pipz.Apply(
		AssignID,
		func(_ context.Context, item *FanOutItem) (*FanOutItem, error) {
			// Deterministic ID from (eventID, subscriptionID) for deduplication.
			// Herald redelivery produces the same ID, making index upserts idempotent.
			item.Notification.ID = uuid.NewSHA1(uuid.NameSpaceOID, []byte(item.EventID+":"+item.Subscription.ID)).String()
			item.Notification.UserID = item.Subscription.UserID
			item.Notification.EventID = item.EventID
			item.Notification.CreatedAt = time.Now()
			item.Notification.Status = models.NotificationUnread
			return item, nil
		},
	)
}
