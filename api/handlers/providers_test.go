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

func TestListProviders_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnListProvidersByTenant: func(_ context.Context, tenantID string, _ models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
			if tenantID != "tenant-1" {
				t.Errorf("expected tenant-1, got %q", tenantID)
			}
			return &models.OffsetResult[models.Provider]{
				Items: []*models.Provider{{ID: "p1", Name: "GDrive", Type: "google_drive", Active: true}},
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/providers", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.ProviderListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Providers) != 1 || resp.Providers[0].ID != "p1" {
		t.Errorf("unexpected providers: %+v", resp.Providers)
	}
}

func TestListProviders_Error(t *testing.T) {
	mock := &argustest.MockProviders{
		OnListProvidersByTenant: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Provider], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/providers", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestCreateProvider_Error(t *testing.T) {
	mock := &argustest.MockProviders{
		OnCreateProvider: func(_ context.Context, _ string, _ models.ProviderType, _, _ string) (*models.Provider, error) {
			return nil, fmt.Errorf("failed")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.ProviderCreateRequest{Type: "s3", Name: "X", Credentials: "c"}
	capture := rtesting.ServeRequest(engine, "POST", "/providers", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestGetProvider_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Name: "GDrive", Type: "google_drive", Active: true}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/providers/p1", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.ProviderResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "p1" || resp.Name != "GDrive" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestGetProvider_NotFound(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/providers/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateProvider_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnCreateProvider: func(_ context.Context, tenantID string, pt models.ProviderType, name, creds string) (*models.Provider, error) {
			return &models.Provider{ID: "p-new", TenantID: tenantID, Type: pt, Name: name, Active: true}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.ProviderCreateRequest{Type: "s3", Name: "My S3", Credentials: "creds"}
	capture := rtesting.ServeRequest(engine, "POST", "/providers", body)
	rtesting.AssertStatus(t, capture, 201)

	var resp wire.ProviderResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "p-new" || resp.Name != "My S3" {
		t.Errorf("unexpected response: %+v", resp)
	}
}

func TestUpdateProvider_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnUpdateProvider: func(_ context.Context, id string, pt models.ProviderType, name, creds string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: pt, Name: name, Active: true}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.ProviderCreateRequest{Type: "s3", Name: "Updated", Credentials: "new-creds"}
	capture := rtesting.ServeRequest(engine, "PUT", "/providers/p1", body)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.ProviderResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Name != "Updated" {
		t.Errorf("expected Updated, got %q", resp.Name)
	}
}

func TestUpdateProvider_NotFound(t *testing.T) {
	mock := &argustest.MockProviders{
		OnUpdateProvider: func(_ context.Context, _ string, _ models.ProviderType, _, _ string) (*models.Provider, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIProviders(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.ProviderCreateRequest{Type: "s3", Name: "X", Credentials: "c"}
	capture := rtesting.ServeRequest(engine, "PUT", "/providers/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}
