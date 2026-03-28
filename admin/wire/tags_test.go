package wire

import "testing"

func TestAdminTagResponse_Clone(t *testing.T) {
	orig := AdminTagResponse{ID: "t1", Name: "compliance"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestAdminTagListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminTagListResponse{
		Tags: []AdminTagResponse{{ID: "t1"}},
	}
	cloned := orig.Clone()

	cloned.Tags[0].ID = "mutated"
	if orig.Tags[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminTagListResponse_Clone_NilSlice(t *testing.T) {
	orig := AdminTagListResponse{}
	cloned := orig.Clone()
	if cloned.Tags != nil {
		t.Error("expected nil Tags")
	}
}
