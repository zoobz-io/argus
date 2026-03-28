//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestTenantToAdminResponse(t *testing.T) {
	tenant := argustest.NewTenant()
	resp := TenantToAdminResponse(tenant)

	if resp.ID != "t1" || resp.Name != "Acme" || resp.Slug != "acme" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if !resp.CreatedAt.Equal(argustest.FixtureTime) || !resp.UpdatedAt.Equal(argustest.FixtureTime2) {
		t.Error("timestamps not mapped")
	}
}

func TestTenantsToAdminList(t *testing.T) {
	result := &models.OffsetResult[models.Tenant]{
		Items:  []*models.Tenant{argustest.NewTenant()},
		Offset: 0,
		Total:  1,
	}

	resp := TenantsToAdminList(result)
	if len(resp.Tenants) != 1 || resp.Tenants[0].ID != "t1" {
		t.Errorf("tenants mismatch: %+v", resp.Tenants)
	}
	if resp.Total != 1 || resp.Offset != 0 {
		t.Errorf("pagination mismatch: %+v", resp)
	}
}
