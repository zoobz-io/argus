//go:build testing

package wire

import (
	"testing"
)

func TestWatchedPathListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := WatchedPathListResponse{
		WatchedPaths: []WatchedPathResponse{{ID: "wp1", Path: "/docs"}},
		Limit:        20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWatchedPathListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := WatchedPathListResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestWatchedPathListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := WatchedPathListResponse{
		WatchedPaths: []WatchedPathResponse{{ID: "wp1"}},
	}
	cloned := orig.Clone()

	cloned.WatchedPaths[0].ID = "mutated"
	if orig.WatchedPaths[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestWatchedPathListResponse_Clone_ValueCopy(t *testing.T) {
	orig := WatchedPathListResponse{Offset: 10, Limit: 20, Total: 100}
	cloned := orig.Clone()

	cloned.Offset = 999
	if orig.Offset == 999 {
		t.Error("mutating clone affected original")
	}
}
