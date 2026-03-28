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

func TestListAllDocumentVersions_Success(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnListDocumentVersions: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
			return &models.OffsetResult[models.DocumentVersion]{Items: []*models.DocumentVersion{{ID: "v1"}}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocumentVersions(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/document-versions", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetAdminDocumentVersion_NotFound(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocumentVersions(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/document-versions/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestListAllDocumentVersions_Error(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnListDocumentVersions: func(_ context.Context, _ models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocumentVersions(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/document-versions", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetAdminDocumentVersion_Success(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnGetDocumentVersion: func(_ context.Context, id string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: id, VersionNumber: 3}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocumentVersions(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/document-versions/v1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestDeleteAdminDocumentVersion_Success(t *testing.T) {
	mock := &argustest.MockDocumentVersions{OnDeleteDocumentVersion: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocumentVersions(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/document-versions/v1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteAdminDocumentVersion_NotFound(t *testing.T) {
	mock := &argustest.MockDocumentVersions{OnDeleteDocumentVersion: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminDocumentVersions(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/document-versions/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
