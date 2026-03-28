package wire

import "testing"

func TestAdminProviderResponse_Clone(t *testing.T) {
	orig := AdminProviderResponse{ID: "p1", Name: "GDrive", Type: "google_drive"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestAdminProviderListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminProviderListResponse{
		Providers: []AdminProviderResponse{{ID: "p1"}},
	}
	cloned := orig.Clone()

	cloned.Providers[0].ID = "mutated"
	if orig.Providers[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminProviderListResponse_Clone_NilFields(t *testing.T) {
	orig := AdminProviderListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Providers != nil {
		t.Error("expected nil Providers")
	}
}
