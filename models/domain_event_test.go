package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDomainEvent_Clone(t *testing.T) {
	evt := DomainEvent{
		ID:           "e-1",
		Timestamp:    time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC),
		Action:       "provider.created",
		ResourceType: "provider",
		ResourceID:   "p-1",
		TenantID:     "t-1",
		ActorID:      "u-1",
		Metadata:     json.RawMessage(`{"provider_type":"google_drive"}`),
	}

	clone := evt.Clone()

	if clone.ID != evt.ID || clone.Action != evt.Action || clone.TenantID != evt.TenantID {
		t.Error("Clone did not copy all fields")
	}

	// Verify deep copy — mutating clone should not affect original.
	clone.Metadata[0] = 'X'
	if evt.Metadata[0] == 'X' {
		t.Error("Clone metadata is not independent")
	}
}

func TestDomainEvent_Clone_NilMetadata(t *testing.T) {
	evt := DomainEvent{ID: "e-2", Action: "tenant.deleted"}
	clone := evt.Clone()
	if clone.Metadata != nil {
		t.Error("nil metadata should remain nil after clone")
	}
}

func TestDomainEventSearchParams_PageSize(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{"zero returns default", 0, DefaultPageSize},
		{"negative returns default", -1, DefaultPageSize},
		{"valid limit", 50, 50},
		{"exceeds max returns max", 200, MaxPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := DomainEventSearchParams{Limit: tt.limit}
			if got := p.PageSize(); got != tt.want {
				t.Errorf("PageSize() = %d, want %d", got, tt.want)
			}
		})
	}
}
