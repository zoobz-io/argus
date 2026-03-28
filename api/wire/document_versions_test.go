//go:build testing

package wire

import (
	"testing"
)

func TestDocumentVersionResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := DocumentVersionResponse{ID: "v1", DocumentID: "d1", VersionNumber: 3}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.VersionNumber != 3 {
		t.Errorf("version number mutated: %d", resp.VersionNumber)
	}
}

func TestDocumentVersionResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := DocumentVersionResponse{ID: "dv1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentVersionResponse_Clone(t *testing.T) {
	orig := DocumentVersionResponse{ID: "v1", DocumentID: "d1", VersionNumber: 3}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}
