//go:build integration

package integration

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/zoobz-io/argus/models"
)

func TestTenants_CreateAndGet(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "Test Corp", "test-corp")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}
	if tenant.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if tenant.Name != "Test Corp" || tenant.Slug != "test-corp" {
		t.Errorf("fields mismatch: %+v", tenant)
	}

	got, err := s.Tenants.GetTenant(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("get tenant: %v", err)
	}
	if got.ID != tenant.ID || got.Name != tenant.Name {
		t.Errorf("get returned different tenant: %+v", got)
	}
}

func TestTenants_Update(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "Original", "original")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	updated, err := s.Tenants.UpdateTenant(ctx, tenant.ID, "Updated", "updated")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Updated" || updated.Slug != "updated" {
		t.Errorf("update fields mismatch: %+v", updated)
	}
}

func TestTenants_Delete(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "ToDelete", "to-delete")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	if err := s.Tenants.DeleteTenant(ctx, tenant.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, err = s.Tenants.GetTenant(ctx, tenant.ID)
	if err == nil {
		t.Fatal("expected error getting deleted tenant")
	}
}

func TestTenants_List(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		_, err := s.Tenants.CreateTenant(ctx, fmt.Sprintf("List Tenant %d", i), fmt.Sprintf("list-tenant-%d-%s", i, uuid.New().String()[:8]))
		if err != nil {
			t.Fatalf("create %d: %v", i, err)
		}
	}

	result, err := s.Tenants.ListTenants(ctx, models.OffsetPage{Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) < 3 {
		t.Errorf("expected at least 3 tenants, got %d", len(result.Items))
	}
}
