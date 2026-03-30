package handlers

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/events"
)

var notificationStream = rocco.NewStreamHandler[rocco.NoBody, wire.NotificationSSE](
	"notification-stream",
	"GET",
	"/notifications/stream",
	func(r *rocco.Request[rocco.NoBody], stream rocco.Stream[wire.NotificationSSE]) error {
		users := sum.MustUse[contracts.Users](r)
		user, err := users.GetUserByExternalID(r, r.Identity.ID())
		if err != nil {
			return stream.SendEvent("error", wire.NotificationSSE{
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

			_ = stream.SendEvent("notification", wire.NotificationSSE{
				NotificationID: hint.NotificationID,
				Type:           hint.Type,
				Message:        hint.Message,
			})
		})
		if listener != nil {
			defer listener.Close()
		}

		// 2. Send connected event.
		if err := stream.SendEvent("connected", wire.NotificationSSE{}); err != nil {
			return err
		}

		// 3. Block until client disconnect.
		<-stream.Done()
		return nil
	},
).WithAuthentication()
