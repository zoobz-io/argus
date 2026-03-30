// Package notify provides the notification fan-out pipeline.
//
// The pipeline processes notification items through assignment, indexing,
// and hint emission stages. Each stage is a pipz processor that transforms
// a FanOutItem as it flows through the sequence.
package notify

import "github.com/zoobz-io/argus/models"

// FanOutItem carries data through the notification fan-out pipeline stages.
type FanOutItem struct {
	Notification       *models.Notification
	Subscription       *models.Subscription
	DomainEvent        *models.DomainEvent // Source event — used by webhook sign stage.
	WebhookHook        *models.Hook        // Loaded by sign stage, reused by deliver stage.
	WebhookDeliveryErr *string
	EventID            string
	WebhookSignature   string
	WebhookTimestamp   string // Unix timestamp included in HMAC computation.
	WebhookPayload     []byte
	WebhookStatusCode  int
}

// Clone returns a deep copy of the fan-out item for pipz compatibility.
func (f *FanOutItem) Clone() *FanOutItem {
	c := &FanOutItem{
		EventID: f.EventID,
	}
	if f.Notification != nil {
		n := f.Notification.Clone()
		c.Notification = &n
	}
	if f.Subscription != nil {
		s := f.Subscription.Clone()
		c.Subscription = &s
	}
	if f.DomainEvent != nil {
		e := f.DomainEvent.Clone()
		c.DomainEvent = &e
	}
	return c
}
