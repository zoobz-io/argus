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

func TestListTags_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) {
			return []*models.Tag{{ID: "tg1", Name: "compliance"}}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/tags", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestListTags_Error(t *testing.T) {
	mock := &argustest.MockTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) { return nil, fmt.Errorf("db error") },
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/tags", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestCreateTag_Error(t *testing.T) {
	mock := &argustest.MockTags{
		OnCreateTag: func(_ context.Context, _, _, _ string) (*models.Tag, error) { return nil, fmt.Errorf("failed") },
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TagCreateRequest{Name: "X", Description: "X desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/tags", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestUpdateTag_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnUpdateTag: func(_ context.Context, id, name, desc string) (*models.Tag, error) {
			return &models.Tag{ID: id, Name: name, Description: desc}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TagCreateRequest{Name: "Updated", Description: "Updated desc"}
	capture := rtesting.ServeRequest(engine, "PUT", "/tags/tg1", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetTag_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnGetTag: func(_ context.Context, id string) (*models.Tag, error) {
			return &models.Tag{ID: id, Name: "compliance", Description: "Compliance docs"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/tags/tg1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetTag_NotFound(t *testing.T) {
	mock := &argustest.MockTags{
		OnGetTag: func(_ context.Context, _ string) (*models.Tag, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/tags/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateTag_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnCreateTag: func(_ context.Context, tenantID, name, desc string) (*models.Tag, error) {
			return &models.Tag{ID: "tg-new", TenantID: tenantID, Name: name, Description: desc}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TagCreateRequest{Name: "security", Description: "Security content"}
	capture := rtesting.ServeRequest(engine, "POST", "/tags", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestUpdateTag_NotFound(t *testing.T) {
	mock := &argustest.MockTags{
		OnUpdateTag: func(_ context.Context, _, _, _ string) (*models.Tag, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TagCreateRequest{Name: "X", Description: "X desc"}
	capture := rtesting.ServeRequest(engine, "PUT", "/tags/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateTag_VocabularyError(t *testing.T) {
	vocab := &argustest.MockVocabulary{
		OnProcess: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("injection detected")
		},
	}
	mock := &argustest.MockTags{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithAPIVocabulary(vocab), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TagCreateRequest{Name: "bad", Description: "bad desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/tags", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status for vocabulary rejection")
	}
}

func TestUpdateTag_VocabularyError(t *testing.T) {
	vocab := &argustest.MockVocabulary{
		OnProcessUpdate: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("injection detected")
		},
	}
	mock := &argustest.MockTags{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITags(mock), argustest.WithAPIVocabulary(vocab), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TagCreateRequest{Name: "bad", Description: "bad desc"}
	capture := rtesting.ServeRequest(engine, "PUT", "/tags/tg1", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status for vocabulary rejection")
	}
}
