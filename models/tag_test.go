package models

import (
	"testing"
	"time"
)

func TestTag_GetID(t *testing.T) {
	tg := Tag{ID: "tag-1"}
	if got := tg.GetID(); got != "tag-1" {
		t.Errorf("GetID() = %q, want %q", got, "tag-1")
	}
}

func TestTag_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tg := Tag{CreatedAt: ts}
	if got := tg.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestTag_Clone(t *testing.T) {
	tg := Tag{ID: "tag-1", Name: "Important", TenantID: "t-1"}
	clone := tg.Clone()

	if clone.ID != tg.ID || clone.Name != tg.Name {
		t.Error("Clone did not copy all fields")
	}

	clone.Name = "mutated"
	if tg.Name != "Important" {
		t.Error("mutating clone leaked to original")
	}
}
