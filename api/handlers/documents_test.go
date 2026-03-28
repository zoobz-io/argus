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

func TestListDocuments_Success(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnListDocumentsByTenant: func(_ context.Context, tenantID string, _ models.OffsetPage) (*models.OffsetResult[models.Document], error) {
			return &models.OffsetResult[models.Document]{
				Items: []*models.Document{{ID: "d1", Name: "report.pdf", MimeType: "application/pdf", ProviderID: "p1"}},
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.DocumentListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Documents) != 1 || resp.Documents[0].ID != "d1" {
		t.Errorf("unexpected: %+v", resp.Documents)
	}
}

func TestListDocuments_Error(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnListDocumentsByTenant: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.Document], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetDocument_Success(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) {
			return &models.Document{ID: id, Name: "report.pdf", MimeType: "application/pdf", ProviderID: "p1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/d1", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.DocumentResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "d1" {
		t.Errorf("expected d1, got %q", resp.ID)
	}
}

func TestGetDocument_NotFound(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocuments(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
