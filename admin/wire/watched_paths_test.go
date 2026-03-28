package wire

import "testing"

func TestAdminWatchedPathResponse_Clone(t *testing.T) {
	orig := AdminWatchedPathResponse{ID: "wp1", Path: "/docs"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestAdminWatchedPathListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminWatchedPathListResponse{
		WatchedPaths: []AdminWatchedPathResponse{{ID: "wp1"}},
	}
	cloned := orig.Clone()

	cloned.WatchedPaths[0].ID = "mutated"
	if orig.WatchedPaths[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminWatchedPathListResponse_Clone_NilFields(t *testing.T) {
	orig := AdminWatchedPathListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.WatchedPaths != nil {
		t.Error("expected nil WatchedPaths")
	}
}
