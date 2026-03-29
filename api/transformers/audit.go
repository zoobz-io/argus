package transformers

import (
	"encoding/json"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// AuditEntryToResponse transforms an AuditEntry model to an API response.
func AuditEntryToResponse(e *models.AuditEntry) wire.AuditEntryResponse {
	var metadata json.RawMessage
	if e.Metadata != "" {
		metadata = json.RawMessage(e.Metadata)
	}
	return wire.AuditEntryResponse{
		ID:           e.ID,
		Timestamp:    e.Timestamp,
		Action:       e.Action,
		ResourceType: e.ResourceType,
		ResourceID:   e.ResourceID,
		ActorID:      e.ActorID,
		Metadata:     metadata,
	}
}

// AuditEntriesToListResponse transforms an OffsetResult of audit entries to a list response.
func AuditEntriesToListResponse(result *models.OffsetResult[models.AuditEntry]) wire.AuditListResponse {
	entries := make([]wire.AuditEntryResponse, len(result.Items))
	for i, e := range result.Items {
		entries[i] = AuditEntryToResponse(e)
	}
	return wire.AuditListResponse{
		Entries: entries,
		Offset:  result.Offset,
		Limit:   len(result.Items),
		Total:   result.Total,
	}
}
