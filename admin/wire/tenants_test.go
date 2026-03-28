package wire

import "testing"

func TestAdminTenantResponse_Clone(t *testing.T) {
	orig := AdminTenantResponse{ID: "t1", Name: "Acme", Slug: "acme"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestAdminTenantListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminTenantListResponse{
		Tenants: []AdminTenantResponse{{ID: "t1", Name: "Acme"}},
		Limit:   20,
	}
	cloned := orig.Clone()

	cloned.Tenants[0].Name = "mutated"
	if orig.Tenants[0].Name == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminTenantListResponse_Clone_NilFields(t *testing.T) {
	orig := AdminTenantListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Tenants != nil {
		t.Error("expected nil Tenants")
	}
}
