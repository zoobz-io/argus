package wire

import (
	"encoding/json"
	"testing"
	"time"
)

func TestProviderCreatedMeta_Clone(t *testing.T) {
	m := ProviderCreatedMeta{ProviderType: "google_drive", ProviderName: "GDrive"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestProviderUpdatedMeta_Clone(t *testing.T) {
	m := ProviderUpdatedMeta{ProviderType: "google_drive", ProviderName: "Updated"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestProviderConnectedMeta_Clone(t *testing.T) {
	m := ProviderConnectedMeta{ProviderType: "google_drive"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestProviderDeletedMeta_Clone(t *testing.T) {
	m := ProviderDeletedMeta{}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestDocumentIngestedMeta_Clone(t *testing.T) {
	m := DocumentIngestedMeta{VersionID: "v-1"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestWatchedPathCreatedMeta_Clone(t *testing.T) {
	m := WatchedPathCreatedMeta{ProviderID: "p-1", Path: "/docs"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestWatchedPathUpdatedMeta_Clone(t *testing.T) {
	m := WatchedPathUpdatedMeta{Path: "/new-path"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTopicCreatedMeta_Clone(t *testing.T) {
	m := TopicCreatedMeta{Name: "Security"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTopicUpdatedMeta_Clone(t *testing.T) {
	m := TopicUpdatedMeta{Name: "Updated"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTagCreatedMeta_Clone(t *testing.T) {
	m := TagCreatedMeta{Name: "compliance"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTagUpdatedMeta_Clone(t *testing.T) {
	m := TagUpdatedMeta{Name: "updated-tag"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTenantCreatedMeta_Clone(t *testing.T) {
	m := TenantCreatedMeta{Name: "Acme", Slug: "acme"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTenantUpdatedMeta_Clone(t *testing.T) {
	m := TenantUpdatedMeta{Name: "Updated", Slug: "updated"}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

func TestTenantDeletedMeta_Clone(t *testing.T) {
	m := TenantDeletedMeta{}
	c := m.Clone()
	if c != m {
		t.Errorf("clone mismatch: %+v vs %+v", c, m)
	}
}

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
