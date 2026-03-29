//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestCreateHook_Success(t *testing.T) {
	users := defaultUsersMock()
	hooks := &argustest.MockHooks{
		OnCreateHook: func(_ context.Context, _, _, url string) (*models.Hook, error) {
			return &models.Hook{
				ID: "hook-1", TenantID: "tenant-1", UserID: "u-1",
				URL: url, Secret: "abc123secret", Active: true,
				CreatedAt: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithAPIHooks(hooks),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	body := wire.HookCreateRequest{URL: "https://example.com/webhook"}
	capture := rtesting.ServeRequest(engine, "POST", "/hooks", body)
	rtesting.AssertStatus(t, capture, 201)

	var resp wire.HookCreateResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "hook-1" {
		t.Errorf("ID = %q, want hook-1", resp.ID)
	}
	if resp.Secret != "abc123secret" {
		t.Errorf("Secret = %q, want abc123secret (should be visible on create)", resp.Secret)
	}
}

func TestCreateHook_UserNotFound(t *testing.T) {
	users := &argustest.MockUsers{
		OnGetUserByExternalID: func(_ context.Context, _ string) (*models.User, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	hooks := &argustest.MockHooks{}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIUsers(users),
		argustest.WithAPIHooks(hooks),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	body := wire.HookCreateRequest{URL: "https://example.com/webhook"}
	capture := rtesting.ServeRequest(engine, "POST", "/hooks", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListHooks_Success(t *testing.T) {
	hooks := &argustest.MockHooks{
		OnListHooksByTenant: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Hook], error) {
			return &models.OffsetResult[models.Hook]{
				Items: []*models.Hook{{ID: "hook-1", URL: "https://example.com", Active: true}},
				Total: 1,
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIHooks(hooks),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/hooks", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestDeleteHook_Success(t *testing.T) {
	hooks := &argustest.MockHooks{
		OnDeleteHook: func(_ context.Context, _, _ string) error {
			return nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIHooks(hooks),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "DELETE", "/hooks/hook-1", nil)
	if capture.StatusCode() >= 400 {
		t.Errorf("expected success, got %d", capture.StatusCode())
	}
}

func TestDeleteHook_NotFound(t *testing.T) {
	hooks := &argustest.MockHooks{
		OnDeleteHook: func(_ context.Context, _, _ string) error {
			return fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIHooks(hooks),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "DELETE", "/hooks/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListDeliveries_Success(t *testing.T) {
	deliveries := &argustest.MockDeliveries{
		OnListByHook: func(_ context.Context, _, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
			return &models.OffsetResult[models.Delivery]{
				Items: []*models.Delivery{{
					ID: "del-1", HookID: "hook-1", EventID: "evt-1",
					StatusCode: 200, Attempt: 1,
				}},
				Total: 1,
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIDeliveries(deliveries),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/hooks/hook-1/deliveries", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestListDeliveries_NotFound(t *testing.T) {
	deliveries := &argustest.MockDeliveries{
		OnListByHook: func(_ context.Context, _, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Delivery], error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIDeliveries(deliveries),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/hooks/missing/deliveries", nil)
	rtesting.AssertStatus(t, capture, 404)
}
