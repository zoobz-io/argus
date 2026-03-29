package notify

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"

	"github.com/zoobz-io/argus/events"
)

func newHintStage() pipz.Chainable[*FanOutItem] {
	return pipz.Effect(
		HintID,
		func(ctx context.Context, item *FanOutItem) error {
			hint := events.NotifyHint{
				UserID:         item.Notification.UserID,
				TenantID:       item.Notification.TenantID,
				NotificationID: item.Notification.ID,
				Type:           string(item.Notification.Type),
				Message:        item.Notification.Message,
			}
			capitan.Info(ctx, events.NotifyHintSignal,
				events.NotifyHintKey.Field(hint),
			)
			return nil
		},
	)
}
