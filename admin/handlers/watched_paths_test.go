//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestListAllWatchedPaths_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnListWatchedPaths: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
			return &models.OffsetResult[models.WatchedPath]{Items: []*models.WatchedPath{{ID: "wp1"}}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminWatchedPaths(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetAdminWatchedPath_NotFound(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnGetWatchedPath: func(_ context.Context, _ string) (*models.WatchedPath, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminWatchedPaths(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListAllWatchedPaths_Error(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnListWatchedPaths: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.WatchedPath], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminWatchedPaths(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetAdminWatchedPath_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{
		OnGetWatchedPath: func(_ context.Context, id string) (*models.WatchedPath, error) {
			return &models.WatchedPath{ID: id, Path: "/docs"}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminWatchedPaths(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/watched-paths/wp1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestDeleteAdminWatchedPath_Success(t *testing.T) {
	mock := &argustest.MockWatchedPaths{OnDeleteWatchedPath: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminWatchedPaths(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/watched-paths/wp1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteAdminWatchedPath_NotFound(t *testing.T) {
	mock := &argustest.MockWatchedPaths{OnDeleteWatchedPath: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminWatchedPaths(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/watched-paths/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
