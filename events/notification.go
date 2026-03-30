package events

import (
	"github.com/zoobz-io/capitan"
)

// NotifyHintSignal is emitted after a notification is indexed, carrying minimal
// data for SSE push to the app. This is a Pub/Sub signal, not a stream.
var NotifyHintSignal = capitan.NewSignal("argus.notify.hint", "Notification hint for fan-out")

// NotifyHintKey carries the NotifyHint payload on the signal.
var NotifyHintKey = capitan.NewKey[NotifyHint]("notify_hint", "events.NotifyHint")

// NotifyHint carries the minimal information needed to fan out a domain event into per-user notifications.
type NotifyHint struct {
	UserID         string
	TenantID       string
	NotificationID string
	Type           string
	Message        string
}

// Notifier sidecar operational signals.
var (
	NotifierIndexed        = capitan.NewSignal("argus.notifier.indexed", "Notification indexed in search")
	NotifierIndexError     = capitan.NewSignal("argus.notifier.index.error", "Failed to index notification")
	NotifierFanOutCompleted = capitan.NewSignal("argus.notifier.fanout.completed", "Notification fan-out completed")
	NotifierFanOutError     = capitan.NewSignal("argus.notifier.fanout.error", "Notification fan-out failed")
)

// Notifier field keys for signal emission.
var (
	NotifierTypeKey  = capitan.NewStringKey("notification_type")
	NotifierErrorKey = capitan.NewErrorKey("error")
)
