package wire

import "testing"

func TestAdminTopicResponse_Clone(t *testing.T) {
	orig := AdminTopicResponse{ID: "t1", Name: "Security"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestAdminTopicListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminTopicListResponse{
		Topics: []AdminTopicResponse{{ID: "t1"}},
	}
	cloned := orig.Clone()

	cloned.Topics[0].ID = "mutated"
	if orig.Topics[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminTopicListResponse_Clone_NilSlice(t *testing.T) {
	orig := AdminTopicListResponse{}
	cloned := orig.Clone()
	if cloned.Topics != nil {
		t.Error("expected nil Topics")
	}
}
