//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestListTenants_Success(t *testing.T) {
	mock := &argustest.MockTenants{
		OnListTenants: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.Tenant], error) {
			return &models.OffsetResult[models.Tenant]{Items: []*models.Tenant{{ID: "t1", Name: "Acme", Slug: "acme"}}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetTenant_Success(t *testing.T) {
	mock := &argustest.MockTenants{
		OnGetTenant: func(_ context.Context, id string) (*models.Tenant, error) {
			return &models.Tenant{ID: id, Name: "Acme", Slug: "acme"}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants/t1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetTenant_NotFound(t *testing.T) {
	mock := &argustest.MockTenants{
		OnGetTenant: func(_ context.Context, _ string) (*models.Tenant, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateTenant_Success(t *testing.T) {
	mock := &argustest.MockTenants{
		OnCreateTenant: func(_ context.Context, name, slug string) (*models.Tenant, error) {
			return &models.Tenant{ID: "t-new", Name: name, Slug: slug}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	body := wire.AdminTenantCreateRequest{Name: "New Corp", Slug: "new-corp"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestUpdateTenant_NotFound(t *testing.T) {
	mock := &argustest.MockTenants{
		OnUpdateTenant: func(_ context.Context, _, _, _ string) (*models.Tenant, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	body := wire.AdminTenantCreateRequest{Name: "X", Slug: "x"}
	capture := rtesting.ServeRequest(engine, "PUT", "/tenants/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListTenants_Error(t *testing.T) {
	mock := &argustest.MockTenants{
		OnListTenants: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.Tenant], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestUpdateTenant_Success(t *testing.T) {
	mock := &argustest.MockTenants{
		OnUpdateTenant: func(_ context.Context, id, name, slug string) (*models.Tenant, error) {
			return &models.Tenant{ID: id, Name: name, Slug: slug}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	body := wire.AdminTenantCreateRequest{Name: "Updated", Slug: "updated"}
	capture := rtesting.ServeRequest(engine, "PUT", "/tenants/t1", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestCreateTenant_Error(t *testing.T) {
	mock := &argustest.MockTenants{
		OnCreateTenant: func(_ context.Context, _, _ string) (*models.Tenant, error) { return nil, fmt.Errorf("duplicate") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	body := wire.AdminTenantCreateRequest{Name: "Dup", Slug: "dup"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestDeleteTenant_Success(t *testing.T) {
	mock := &argustest.MockTenants{OnDeleteTenant: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/tenants/t1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteTenant_NotFound(t *testing.T) {
	mock := &argustest.MockTenants{OnDeleteTenant: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTenants(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/tenants/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
