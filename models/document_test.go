package models

import (
	"testing"
	"time"
)

func TestDocument_GetID(t *testing.T) {
	d := Document{ID: "doc-1"}
	if got := d.GetID(); got != "doc-1" {
		t.Errorf("GetID() = %q, want %q", got, "doc-1")
	}
}

func TestDocument_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	d := Document{CreatedAt: ts}
	if got := d.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestDocument_Clone(t *testing.T) {
	cvid := "ver-current"
	d := Document{
		ID:               "doc-1",
		Name:             "report.pdf",
		CurrentVersionID: &cvid,
	}

	clone := d.Clone()

	// Fields copied.
	if clone.ID != d.ID {
		t.Errorf("ID: got %q, want %q", clone.ID, d.ID)
	}
	if clone.Name != d.Name {
		t.Errorf("Name: got %q, want %q", clone.Name, d.Name)
	}

	// Pointer independence.
	if clone.CurrentVersionID == d.CurrentVersionID {
		t.Error("CurrentVersionID: clone shares pointer with original")
	}
	if *clone.CurrentVersionID != cvid {
		t.Errorf("CurrentVersionID value: got %q, want %q", *clone.CurrentVersionID, cvid)
	}

	// Mutate clone, verify original unchanged.
	*clone.CurrentVersionID = "mutated"
	if *d.CurrentVersionID != cvid {
		t.Error("mutating clone leaked to original")
	}
}

func TestDocument_Clone_NilPointer(t *testing.T) {
	d := Document{ID: "doc-1"}
	clone := d.Clone()
	if clone.CurrentVersionID != nil {
		t.Error("expected nil CurrentVersionID in clone")
	}
}
