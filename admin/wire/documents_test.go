package wire

import "testing"

func TestAdminDocumentResponse_Clone_PointerIndependence(t *testing.T) {
	vid := "v1"
	orig := AdminDocumentResponse{ID: "d1", CurrentVersionID: &vid}
	cloned := orig.Clone()

	if cloned.CurrentVersionID == orig.CurrentVersionID {
		t.Error("pointer not deep copied")
	}
	*cloned.CurrentVersionID = "mutated"
	if *orig.CurrentVersionID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminDocumentResponse_Clone_NilPointer(t *testing.T) {
	orig := AdminDocumentResponse{ID: "d1"}
	cloned := orig.Clone()
	if cloned.CurrentVersionID != nil {
		t.Error("expected nil CurrentVersionID")
	}
}

func TestAdminDocumentListResponse_Clone_SliceIndependence(t *testing.T) {
	vid := "v1"
	orig := AdminDocumentListResponse{
		Documents: []AdminDocumentResponse{{ID: "d1", CurrentVersionID: &vid}},
	}
	cloned := orig.Clone()

	cloned.Documents[0].ID = "mutated"
	if orig.Documents[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminDocumentListResponse_Clone_NilFields(t *testing.T) {
	orig := AdminDocumentListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Documents != nil {
		t.Error("expected nil Documents")
	}
}
