//go:build testing

package wire

import (
	"testing"
)

func TestDocumentListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := DocumentListResponse{
		Documents: []DocumentResponse{{ID: "d1", Name: "a.pdf"}},
		Limit:     20,
		Total:     100,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Documents) != 1 || resp.Documents[0].ID != "d1" {
		t.Errorf("documents mutated: %+v", resp.Documents)
	}
}

func TestDocumentListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := DocumentListResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentListResponse_Clone_SliceIndependence(t *testing.T) {
	vid := "v1"
	orig := DocumentListResponse{
		Documents: []DocumentResponse{{ID: "d1", CurrentVersionID: &vid}},
		Limit:     20,
	}
	cloned := orig.Clone()

	cloned.Documents[0].ID = "mutated"
	if orig.Documents[0].ID == "mutated" {
		t.Error("mutating clone slice affected original")
	}
}

func TestDocumentListResponse_Clone_NilFields(t *testing.T) {
	orig := DocumentListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Documents != nil {
		t.Error("expected nil Documents")
	}
}
