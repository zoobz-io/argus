package models

import (
	"testing"
	"time"
)

func TestDocumentVersion_GetID(t *testing.T) {
	v := DocumentVersion{ID: "ver-1"}
	if got := v.GetID(); got != "ver-1" {
		t.Errorf("GetID() = %q, want %q", got, "ver-1")
	}
}

func TestDocumentVersion_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	v := DocumentVersion{CreatedAt: ts}
	if got := v.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestDocumentVersion_Clone(t *testing.T) {
	v := DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1", VersionNumber: 3}
	clone := v.Clone()

	if clone.ID != v.ID || clone.DocumentID != v.DocumentID || clone.VersionNumber != v.VersionNumber {
		t.Error("Clone did not copy all fields")
	}

	clone.ID = "mutated"
	if v.ID != "ver-1" {
		t.Error("mutating clone leaked to original")
	}
}
