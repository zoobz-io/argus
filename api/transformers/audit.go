package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// DomainEventToResponse transforms a DomainEvent model to an API response.
func DomainEventToResponse(e *models.DomainEvent) wire.AuditEntryResponse {
	return wire.AuditEntryResponse{
		ID:           e.ID,
		Timestamp:    e.Timestamp,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		ActorID:      e.ActorID,
		Metadata:     e.Metadata,
	}
}

// DomainEventsToListResponse transforms an OffsetResult of domain events to a list response.
func DomainEventsToListResponse(result *models.OffsetResult[models.DomainEvent]) wire.AuditListResponse {
	entries := make([]wire.AuditEntryResponse, len(result.Items))
	for i, e := range result.Items {
		entries[i] = DomainEventToResponse(e)
	}
	return wire.AuditListResponse{
		Entries: entries,
		Offset:  result.Offset,
		Limit:   len(result.Items),
		Total:   result.Total,
	}
}
