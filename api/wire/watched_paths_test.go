//go:build testing

package wire

import (
	"testing"
)

func TestWatchedPathCreateRequest_Validate_Valid(t *testing.T) {
	req := WatchedPathCreateRequest{Path: "/documents/reports"}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestWatchedPathCreateRequest_Validate_MissingPath(t *testing.T) {
	req := WatchedPathCreateRequest{}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty path")
	}
}

func TestWatchedPathCreateRequest_Clone(t *testing.T) {
	orig := WatchedPathCreateRequest{Path: "/docs", ProviderID: "p1"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestWatchedPathResponse_Clone(t *testing.T) {
	orig := WatchedPathResponse{ID: "wp1", Path: "/docs", ProviderID: "p1", Active: true}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestWatchedPathResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := WatchedPathResponse{ID: "wp1", Path: "/docs", ProviderID: "p1", Active: true}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Path != "/docs" {
		t.Errorf("Path mutated: %q", resp.Path)
	}
}

func TestWatchedPathResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := WatchedPathResponse{ID: "wp1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}
