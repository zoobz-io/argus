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

func TestListUsers_Success(t *testing.T) {
	mock := &argustest.MockUsers{
		OnListUsers: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.User], error) {
			return &models.OffsetResult[models.User]{Items: []*models.User{{ID: "u1", Email: "alice@example.com", DisplayName: "Alice", Role: models.UserRoleViewer, Status: models.UserStatusActive}}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/users", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestListUsers_Error(t *testing.T) {
	mock := &argustest.MockUsers{
		OnListUsers: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.User], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/users", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetUser_Success(t *testing.T) {
	mock := &argustest.MockUsers{
		OnGetUser: func(_ context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, Email: "alice@example.com", DisplayName: "Alice", Role: models.UserRoleViewer, Status: models.UserStatusActive}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/users/u1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetUser_NotFound(t *testing.T) {
	mock := &argustest.MockUsers{
		OnGetUser: func(_ context.Context, _ string) (*models.User, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/users/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateUser_Success(t *testing.T) {
	mock := &argustest.MockUsers{
		OnCreateUser: func(_ context.Context, tenantID, externalID, email, displayName string, role models.UserRole) (*models.User, error) {
			return &models.User{ID: "u-new", TenantID: tenantID, ExternalID: externalID, Email: email, DisplayName: displayName, Role: role, Status: models.UserStatusActive}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	body := wire.AdminUserCreateRequest{TenantID: "t-1", ExternalID: "ext-1", Email: "alice@example.com", DisplayName: "Alice", Role: models.UserRoleViewer}
	capture := rtesting.ServeRequest(engine, "POST", "/users", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestCreateUser_Error(t *testing.T) {
	mock := &argustest.MockUsers{
		OnCreateUser: func(_ context.Context, _, _, _, _ string, _ models.UserRole) (*models.User, error) {
			return nil, fmt.Errorf("duplicate")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	body := wire.AdminUserCreateRequest{TenantID: "t-1", ExternalID: "ext-1", Email: "alice@example.com", DisplayName: "Alice", Role: models.UserRoleViewer}
	capture := rtesting.ServeRequest(engine, "POST", "/users", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestUpdateUser_Success(t *testing.T) {
	mock := &argustest.MockUsers{
		OnGetUser: func(_ context.Context, id string) (*models.User, error) {
			return &models.User{ID: id, Email: "old@example.com", DisplayName: "Old", Role: models.UserRoleViewer, Status: models.UserStatusActive}, nil
		},
		OnUpdateUser: func(_ context.Context, id, email, displayName string, role models.UserRole, status models.UserStatus) (*models.User, error) {
			return &models.User{ID: id, Email: email, DisplayName: displayName, Role: role, Status: status}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	newEmail := "new@example.com"
	body := wire.AdminUserUpdateRequest{Email: &newEmail}
	capture := rtesting.ServeRequest(engine, "PUT", "/users/u1", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestUpdateUser_NotFound(t *testing.T) {
	mock := &argustest.MockUsers{
		OnGetUser: func(_ context.Context, _ string) (*models.User, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	newEmail := "new@example.com"
	body := wire.AdminUserUpdateRequest{Email: &newEmail}
	capture := rtesting.ServeRequest(engine, "PUT", "/users/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestDeleteUser_Success(t *testing.T) {
	mock := &argustest.MockUsers{OnDeleteUser: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/users/u1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteUser_NotFound(t *testing.T) {
	mock := &argustest.MockUsers{OnDeleteUser: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminUsers(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/users/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
