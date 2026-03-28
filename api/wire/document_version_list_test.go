//go:build testing

package wire

import (
	"testing"
)

func TestDocumentVersionListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := DocumentVersionListResponse{
		Versions: []DocumentVersionResponse{{ID: "v1", VersionNumber: 1}},
		Limit:    20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocumentVersionListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := DocumentVersionListResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentVersionListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := DocumentVersionListResponse{
		Versions: []DocumentVersionResponse{{ID: "v1"}},
	}
	cloned := orig.Clone()

	cloned.Versions[0].ID = "mutated"
	if orig.Versions[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestDocumentVersionListResponse_Clone_ValueCopy(t *testing.T) {
	orig := DocumentVersionListResponse{Offset: 10, Limit: 20, Total: 100}
	cloned := orig.Clone()

	cloned.Offset = 999
	if orig.Offset == 999 {
		t.Error("mutating clone affected original")
	}
}
