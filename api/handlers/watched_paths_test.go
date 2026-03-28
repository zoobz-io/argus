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

func TestListWatchedPaths_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnListWatchedPathsByTenant: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
			return &models.OffsetResult[models.WatchedPath]{
				Items: []*models.WatchedPath{{ID: "wp1", Path: "/docs", ProviderID: "p1", Active: true}},
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestListWatchedPaths_Error(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnListWatchedPathsByTenant: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestCreateWatchedPath_Error(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnCreateWatchedPath: func(_ context.Context, _, _, _ string) (*models.WatchedPath, error) {
			return nil, fmt.Errorf("failed")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.WatchedPathCreateRequest{Path: "/x", ProviderID: "p1"}
	capture := rtesting.ServeRequest(engine, "POST", "/watched-paths", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestUpdateWatchedPath_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnUpdateWatchedPath: func(_ context.Context, id, path string) (*models.WatchedPath, error) {
			return &models.WatchedPath{ID: id, Path: path, Active: true}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.WatchedPathCreateRequest{Path: "/updated"}
	capture := rtesting.ServeRequest(engine, "PUT", "/watched-paths/wp1", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetWatchedPath_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnGetWatchedPath: func(_ context.Context, id string) (*models.WatchedPath, error) {
			return &models.WatchedPath{ID: id, Path: "/docs", ProviderID: "p1", Active: true}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths/wp1", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.WatchedPathResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "wp1" || resp.Path != "/docs" {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestGetWatchedPath_NotFound(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnGetWatchedPath: func(_ context.Context, _ string) (*models.WatchedPath, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateWatchedPath_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnCreateWatchedPath: func(_ context.Context, _, providerID, path string) (*models.WatchedPath, error) {
			return &models.WatchedPath{ID: "wp-new", ProviderID: providerID, Path: path, Active: true}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.WatchedPathCreateRequest{Path: "/new/path", ProviderID: "p1"}
	capture := rtesting.ServeRequest(engine, "POST", "/watched-paths", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestUpdateWatchedPath_NotFound(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnUpdateWatchedPath: func(_ context.Context, _, _ string) (*models.WatchedPath, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIWatchedPaths(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.WatchedPathCreateRequest{Path: "/updated"}
	capture := rtesting.ServeRequest(engine, "PUT", "/watched-paths/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}
