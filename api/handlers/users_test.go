//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestListUsersInTenant_Success(t *testing.T) {
	mock := &argustest.MockUsers{
		OnListUsersByTenant: func(_ context.Context, tenantID string, _ models.OffsetPage) (*models.OffsetResult[models.User], error) {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %q", tenantID)
			}
			return &models.OffsetResult[models.User]{
				Items: []*models.User{{ID: "u1", Email: "alice@example.com", DisplayName: "Alice", Role: models.UserRoleViewer}},
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/users", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.UserListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Users) != 1 || resp.Users[0].ID != "u1" {
		t.Errorf("unexpected users: %+v", resp.Users)
	}
}

func TestListUsersInTenant_Error(t *testing.T) {
	mock := &argustest.MockUsers{
		OnListUsersByTenant: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.User], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/users", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetMyProfile_Success(t *testing.T) {
	mock := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, externalID string) (*models.User, error) {
			return &models.User{ID: "u1", ExternalID: externalID, Email: "alice@example.com", DisplayName: "Alice", Role: models.UserRoleViewer}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/users/me", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.UserResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "u1" {
		t.Errorf("unexpected ID: %q", resp.ID)
	}
}

func TestGetMyProfile_NotFound(t *testing.T) {
	mock := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIUsers(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/users/me", nil)
	rtesting.AssertStatus(t, capture, 404)
}
