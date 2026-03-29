package wire

import (
	"encoding/json"
	"testing"
	"time"
)

func TestAuditEntryResponse_Clone(t *testing.T) {
	r := AuditEntryResponse{
		ID:           "a-1",
		Timestamp:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Action:       "provider.created",
		ResourceType: "provider",
		ResourceID:   "p-1",
		ActorID:      "u-1",
		Metadata:     json.RawMessage(`{"key":"value"}`),
	}

	clone := r.Clone()
	if clone.ID != r.ID || clone.Action != r.Action {
		t.Error("Clone did not copy all fields")
	}

	// Mutate clone metadata — original must be unaffected.
	clone.Metadata[0] = 'X'
	if r.Metadata[0] == 'X' {
		t.Error("mutating clone metadata leaked to original")
	}
}

func TestAuditEntryResponse_Clone_NilMetadata(t *testing.T) {
	r := AuditEntryResponse{ID: "a-2"}
	clone := r.Clone()
	if clone.Metadata != nil {
		t.Error("nil metadata should remain nil after clone")
	}
}

func TestAuditListResponse_Clone(t *testing.T) {
	r := AuditListResponse{
		Entries: []AuditEntryResponse{
			{ID: "a-1", Action: "provider.created"},
			{ID: "a-2", Action: "tenant.created"},
		},
		Offset: 0,
		Limit:  2,
		Total:  10,
	}

	clone := r.Clone()
	if len(clone.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(clone.Entries))
	}

	clone.Entries[0].Action = "mutated"
	if r.Entries[0].Action != "provider.created" {
		t.Error("mutating clone entries leaked to original")
	}
}

func TestAuditListResponse_Clone_NilEntries(t *testing.T) {
	r := AuditListResponse{}
	clone := r.Clone()
	if clone.Entries != nil {
		t.Error("nil entries should remain nil after clone")
	}
}
