package models

import (
	"testing"
	"time"
)

func TestTopic_GetID(t *testing.T) {
	tp := Topic{ID: "topic-1"}
	if got := tp.GetID(); got != "topic-1" {
		t.Errorf("GetID() = %q, want %q", got, "topic-1")
	}
}

func TestTopic_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	tp := Topic{CreatedAt: ts}
	if got := tp.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestTopic_Clone(t *testing.T) {
	tp := Topic{ID: "topic-1", Name: "Finance", TenantID: "t-1"}
	clone := tp.Clone()

	if clone.ID != tp.ID || clone.Name != tp.Name {
		t.Error("Clone did not copy all fields")
	}

	clone.Name = "mutated"
	if tp.Name != "Finance" {
		t.Error("mutating clone leaked to original")
	}
}
