//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestTenantToResponse(t *testing.T) {
	tenant := argustest.NewTenant()
	resp := TenantToResponse(tenant)

	if resp.ID != "t1" || resp.Name != "Acme" || resp.Slug != "acme" {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestTenantsToResponse(t *testing.T) {
	tenants := []*models.Tenant{argustest.NewTenant(), argustest.NewTenant()}
	tenants[1].ID = "t2"

	result := TenantsToResponse(tenants)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].ID != "t1" || result[1].ID != "t2" {
		t.Errorf("IDs mismatch: %q, %q", result[0].ID, result[1].ID)
	}
}

func TestTenantsToResponse_Empty(t *testing.T) {
	result := TenantsToResponse(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}
