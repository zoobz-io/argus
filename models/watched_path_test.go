package models

import (
	"testing"
	"time"
)

func TestWatchedPath_GetID(t *testing.T) {
	w := WatchedPath{ID: "wp-1"}
	if got := w.GetID(); got != "wp-1" {
		t.Errorf("GetID() = %q, want %q", got, "wp-1")
	}
}

func TestWatchedPath_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	w := WatchedPath{CreatedAt: ts}
	if got := w.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestWatchedPath_Clone(t *testing.T) {
	w := WatchedPath{ID: "wp-1", Path: "/docs", Active: true}
	clone := w.Clone()

	if clone.ID != w.ID || clone.Path != w.Path || clone.Active != w.Active {
		t.Error("Clone did not copy all fields")
	}

	clone.Path = "mutated"
	if w.Path != "/docs" {
		t.Error("mutating clone leaked to original")
	}
}
