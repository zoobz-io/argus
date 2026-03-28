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

func TestListDocumentVersions_Success(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnListVersionsByDocument: func(_ context.Context, docID string, _ models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
			return &models.OffsetResult[models.DocumentVersion]{
				Items: []*models.DocumentVersion{{ID: "v1", DocumentID: docID, VersionNumber: 1}},
			}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersions(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/d1/versions", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.DocumentVersionListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Versions) != 1 {
		t.Errorf("expected 1 version, got %d", len(resp.Versions))
	}
}

func TestListDocumentVersions_Error(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnListVersionsByDocument: func(_ context.Context, _ string, _ models.OffsetPage) (*models.OffsetResult[models.DocumentVersion], error) {
			return nil, fmt.Errorf("db error")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersions(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/d1/versions", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetDocumentVersion_Success(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnGetDocumentVersion: func(_ context.Context, id string) (*models.DocumentVersion, error) {
			return &models.DocumentVersion{ID: id, DocumentID: "d1", VersionNumber: 3}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersions(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/d1/versions/v1", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.DocumentVersionResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "v1" || resp.VersionNumber != 3 {
		t.Errorf("unexpected: %+v", resp)
	}
}

func TestGetDocumentVersion_NotFound(t *testing.T) {
	mock := &argustest.MockDocumentVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIDocumentVersions(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/documents/d1/versions/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}
