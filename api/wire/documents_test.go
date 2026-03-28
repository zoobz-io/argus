//go:build testing

package wire

import (
	"testing"
)

func TestDocumentResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	vid := "v1"
	resp := DocumentResponse{ID: "d1", Name: "report.pdf", MimeType: "application/pdf", ProviderID: "p1", CurrentVersionID: &vid}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "d1" || resp.Name != "report.pdf" {
		t.Errorf("fields mutated unexpectedly: %+v", resp)
	}
}

func TestDocumentResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := DocumentResponse{ID: "d1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentResponse_Clone_PointerIndependence(t *testing.T) {
	vid := "v1"
	orig := DocumentResponse{ID: "d1", CurrentVersionID: &vid}
	cloned := orig.Clone()

	if cloned.CurrentVersionID == orig.CurrentVersionID {
		t.Error("pointer not deep copied")
	}
	if *cloned.CurrentVersionID != "v1" {
		t.Errorf("value mismatch: %q", *cloned.CurrentVersionID)
	}

	*cloned.CurrentVersionID = "mutated"
	if *orig.CurrentVersionID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestDocumentResponse_Clone_NilPointer(t *testing.T) {
	orig := DocumentResponse{ID: "d1"}
	cloned := orig.Clone()
	if cloned.CurrentVersionID != nil {
		t.Error("expected nil CurrentVersionID")
	}
}
