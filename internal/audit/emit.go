// Package audit provides helpers for emitting audit log entries from handlers.
package audit

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
)

// Emit publishes an audit entry onto the AuditSignal for downstream indexing.
func Emit(ctx context.Context, action, resourceType, resourceID, tenantID, actorID string, metadata map[string]any) {
	var metaStr string
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			metaStr = string(b)
		}
	}
	entry := models.AuditEntry{
		ID:           uuid.NewString(),
		Timestamp:    time.Now().UTC(),
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		TenantID:     tenantID,
		ActorID:      actorID,
		Metadata:     metaStr,
	}
	capitan.Info(ctx, events.AuditSignal, events.AuditKey.Field(entry))
}
