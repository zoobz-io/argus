//go:build testing

package transformers

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
)

func TestDomainEventToResponse(t *testing.T) {
	ts := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	e := &models.DomainEvent{
		ID:           "a-1",
		Timestamp:    ts,
		Action:       "provider.created",
		ResourceType: "provider",
		ResourceID:   "p-1",
		ActorID:      "u-1",
		TenantID:     "t-1",
		Metadata:     json.RawMessage(`{"provider_type":"google_drive"}`),
	}

	resp := DomainEventToResponse(e)

	if resp.ID != "a-1" {
		t.Errorf("ID = %q, want a-1", resp.ID)
	}
	if resp.Action != "provider.created" {
		t.Errorf("Action = %q, want provider.created", resp.Action)
	}
	if resp.ResourceType != "provider" {
		t.Errorf("ResourceType = %q, want provider", resp.ResourceType)
	}
	if resp.ResourceID != "p-1" {
		t.Errorf("ResourceID = %q, want p-1", resp.ResourceID)
	}
	if resp.ActorID != "u-1" {
		t.Errorf("ActorID = %q, want u-1", resp.ActorID)
	}
	if !resp.Timestamp.Equal(ts) {
		t.Errorf("Timestamp = %v, want %v", resp.Timestamp, ts)
	}
	if string(resp.Metadata) != `{"provider_type":"google_drive"}` {
		t.Errorf("Metadata = %s, want provider_type JSON", string(resp.Metadata))
	}
}

func TestDomainEventsToListResponse(t *testing.T) {
	e1 := &models.DomainEvent{ID: "a-1", Action: "provider.created"}
	e2 := &models.DomainEvent{ID: "a-2", Action: "tenant.created"}

	result := &models.OffsetResult[models.DomainEvent]{
		Items:  []*models.DomainEvent{e1, e2},
		Total:  5,
		Offset: 10,
	}

	resp := DomainEventsToListResponse(result)
	if len(resp.Entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp.Entries))
	}
	if resp.Entries[0].ID != "a-1" || resp.Entries[1].ID != "a-2" {
		t.Errorf("IDs mismatch: %q, %q", resp.Entries[0].ID, resp.Entries[1].ID)
	}
	if resp.Total != 5 {
		t.Errorf("Total = %d, want 5", resp.Total)
	}
	if resp.Offset != 10 {
		t.Errorf("Offset = %d, want 10", resp.Offset)
	}
	if resp.Limit != 2 {
		t.Errorf("Limit = %d, want 2", resp.Limit)
	}
}

func TestDomainEventsToListResponse_Empty(t *testing.T) {
	result := &models.OffsetResult[models.DomainEvent]{
		Items: []*models.DomainEvent{},
		Total: 0,
	}
	resp := DomainEventsToListResponse(result)
	if len(resp.Entries) != 0 {
		t.Errorf("expected empty, got %d", len(resp.Entries))
	}
}
