package events

import (
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/models"
)

// DomainEventSignal is emitted when any domain action occurs (user or system).
// Herald publishes this to the unified events stream for the notifier sidecar.
var DomainEventSignal = capitan.NewSignal("argus.domain_event", "Domain event")

// DomainEventKey carries the DomainEvent payload on the signal.
var DomainEventKey = capitan.NewKey[models.DomainEvent]("domain_event", "models.DomainEvent")

// Domain event sidecar operational signals.
var (
	DomainEventIndexed    = capitan.NewSignal("argus.domain_event.indexed", "Domain event indexed in search")
	DomainEventIndexError = capitan.NewSignal("argus.domain_event.index.error", "Failed to index domain event")
)

// Domain event field keys for signal emission.
var (
	DomainEventActionKey = capitan.NewStringKey("domain_event_action")
	DomainEventErrorKey  = capitan.NewErrorKey("error")
)
