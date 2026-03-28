package events

import (
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/models"
)

// NotificationSignal is emitted when a domain event should become a user notification.
// Herald publishes this to the notification stream for the sidecar to consume.
var NotificationSignal = capitan.NewSignal("argus.notification", "Application notification")

// NotificationKey carries the pre-built Notification payload on the signal.
var NotificationKey = capitan.NewKey[models.Notification]("notification", "models.Notification")

// Notifier sidecar operational signals.
var (
	NotifierIndexed    = capitan.NewSignal("argus.notifier.indexed", "Notification indexed in search")
	NotifierIndexError = capitan.NewSignal("argus.notifier.index.error", "Failed to index notification")
)

// Notifier field keys for signal emission.
var (
	NotifierTypeKey  = capitan.NewStringKey("notification_type")
	NotifierErrorKey = capitan.NewErrorKey("error")
)
