package wire

import "testing"

func TestAdminDocumentVersionResponse_Clone(t *testing.T) {
	orig := AdminDocumentVersionResponse{ID: "v1", DocumentID: "d1", VersionNumber: 3}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestAdminDocumentVersionListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminDocumentVersionListResponse{
		Versions: []AdminDocumentVersionResponse{{ID: "v1"}},
	}
	cloned := orig.Clone()

	cloned.Versions[0].ID = "mutated"
	if orig.Versions[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminDocumentVersionListResponse_Clone_NilFields(t *testing.T) {
	orig := AdminDocumentVersionListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Versions != nil {
		t.Error("expected nil Versions")
	}
}
