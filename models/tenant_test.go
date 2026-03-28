package models

import (
	"testing"
	"time"
)

func TestTenant_GetID(t *testing.T) {
	tn := Tenant{ID: "t-1"}
	if got := tn.GetID(); got != "t-1" {
		t.Errorf("GetID() = %q, want %q", got, "t-1")
	}
}

func TestTenant_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tn := Tenant{CreatedAt: ts}
	if got := tn.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestTenant_Clone(t *testing.T) {
	tn := Tenant{ID: "t-1", Name: "Acme", Slug: "acme"}
	clone := tn.Clone()

	if clone.ID != tn.ID || clone.Name != tn.Name || clone.Slug != tn.Slug {
		t.Error("Clone did not copy all fields")
	}

	clone.Name = "mutated"
	if tn.Name != "Acme" {
		t.Error("mutating clone leaked to original")
	}
}
