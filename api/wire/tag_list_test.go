//go:build testing

package wire

import (
	"testing"
)

func TestTagListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := TagListResponse{
		Tags: []TagResponse{{ID: "t1", Name: "compliance"}},
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTagListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := TagListResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestTagListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := TagListResponse{
		Tags: []TagResponse{{ID: "t1"}},
	}
	cloned := orig.Clone()

	cloned.Tags[0].ID = "mutated"
	if orig.Tags[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestTagListResponse_Clone_NilSlice(t *testing.T) {
	orig := TagListResponse{}
	cloned := orig.Clone()
	if cloned.Tags != nil {
		t.Error("expected nil Tags")
	}
}
