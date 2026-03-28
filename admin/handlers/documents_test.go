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

func TestListAllDocuments_Success(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnListDocuments: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.Document], error) {
			return &models.OffsetResult[models.Document]{Items: []*models.Document{{ID: "d1"}}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocuments(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/documents", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetAdminDocument_NotFound(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, _ string) (*models.Document, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocuments(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListAllDocuments_Error(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnListDocuments: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.Document], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocuments(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/documents", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetAdminDocument_Success(t *testing.T) {
	mock := &argustest.MockDocuments{
		OnGetDocument: func(_ context.Context, id string) (*models.Document, error) {
			return &models.Document{ID: id, Name: "report.pdf"}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocuments(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/d1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestDeleteAdminDocument_Success(t *testing.T) {
	mock := &argustest.MockDocuments{OnDeleteDocument: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocuments(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/documents/d1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteAdminDocument_NotFound(t *testing.T) {
	mock := &argustest.MockDocuments{OnDeleteDocument: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocuments(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/documents/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
