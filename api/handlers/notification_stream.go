package handlers

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/events"
)

// NotificationSSE is the SSE event payload sent to clients for real-time notifications.
type NotificationSSE struct {
	NotificationID string `json:"notification_id"`
	Type           string `json:"type"`
	Message        string `json:"message"`
}

// Clone returns a copy of the event.
func (e NotificationSSE) Clone() NotificationSSE {
	return e
}

var notificationStream = rocco.NewStreamHandler[rocco.NoBody, NotificationSSE](
	"notification-stream",
	"GET",
	"/notifications/stream",
	func(r *rocco.Request[rocco.NoBody], stream rocco.Stream[NotificationSSE]) error {
		users := sum.MustUse[contracts.Users](r)
		user, err := users.GetUserByExternalID(r, r.Identity.ID())
		if err != nil {
			return stream.SendEvent("error", NotificationSSE{
				Message: "user not found",
			})
		}

		userID := user.ID
		tid := tenantID(r.Identity)

		// 1. Subscribe to live hint signals BEFORE sending connected event.
		listener := capitan.Hook(events.NotifyHintSignal, func(_ context.Context, e *capitan.Event) {
			hint, ok := events.NotifyHintKey.From(e)
			if !ok || hint.UserID != userID || hint.TenantID != tid {
				return
			}

			_ = stream.SendEvent("notification", NotificationSSE{
				NotificationID: hint.NotificationID,
				Type:           hint.Type,
				Message:        hint.Message,
			})
		})
		if listener != nil {
			defer listener.Close()
		}

		// 2. Send connected event.
		if err := stream.SendEvent("connected", NotificationSSE{}); err != nil {
			return err
		}

		// 3. Block until client disconnect.
		<-stream.Done()
		return nil
	},
).WithAuthentication()
