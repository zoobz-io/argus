// Package event provides helpers for emitting domain events from handlers and pipelines.
package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
)

// Emit publishes a domain event onto the DomainEventSignal for downstream
// indexing, notification fan-out, and webhook delivery. The message is a
// human-readable description of the event (typically from the capitan signal description).
func Emit(ctx context.Context, action, message, resourceType, resourceID, tenantID, actorID string, metadata map[string]any) {
	var metaJSON json.RawMessage
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			metaJSON = b
		}
	}
	evt := models.DomainEvent{
		ID:           uuid.NewString(),
		Timestamp:    time.Now().UTC(),
		Action:       action,
		Message:      message,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     tenantID,
		ActorID:      actorID,
		Metadata:     metaJSON,
	}
	capitan.Info(ctx, events.DomainEventSignal, events.DomainEventKey.Field(evt))
}
