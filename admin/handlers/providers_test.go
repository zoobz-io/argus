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

func TestListAllProviders_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnListProviders: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
			return &models.OffsetResult[models.Provider]{Items: []*models.Provider{{ID: "p1"}}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/providers", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetAdminProvider_NotFound(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProvider: func(_ context.Context, _ string) (*models.Provider, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/providers/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateAdminProvider_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnCreateProvider: func(_ context.Context, tid string, pt models.ProviderType, name, creds string) (*models.Provider, error) {
			return &models.Provider{ID: "p-new", TenantID: tid, Type: pt, Name: name}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	body := wire.ProviderCreateRequest{Type: "s3", Name: "My S3", Credentials: "creds"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/providers", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestListAllProviders_Error(t *testing.T) {
	mock := &argustest.MockProviders{
		OnListProviders: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/providers", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetAdminProvider_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Name: "GDrive"}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/providers/p1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestCreateAdminProvider_Error(t *testing.T) {
	mock := &argustest.MockProviders{
		OnCreateProvider: func(_ context.Context, _ string, _ models.ProviderType, _, _ string) (*models.Provider, error) {
			return nil, fmt.Errorf("failed")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	body := wire.ProviderCreateRequest{Type: "s3", Name: "X", Credentials: "c"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/providers", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestDeleteAdminProvider_Success(t *testing.T) {
	mock := &argustest.MockProviders{OnDeleteProvider: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/providers/p1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteAdminProvider_NotFound(t *testing.T) {
	mock := &argustest.MockProviders{OnDeleteProvider: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminProviders(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/providers/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
