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

func TestListAdminTags_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) {
			return []*models.Tag{{ID: "tg1", Name: "compliance"}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants/t1/tags", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetAdminTag_NotFound(t *testing.T) {
	mock := &argustest.MockTags{
		OnGetTag: func(_ context.Context, _ string) (*models.Tag, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tags/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateAdminTag_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnCreateTag: func(_ context.Context, tid, name, desc string) (*models.Tag, error) {
			return &models.Tag{ID: "tg-new", TenantID: tid, Name: name, Description: desc}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock), argustest.WithAdminVocabulary(&argustest.MockVocabulary{}))
	body := wire.TagCreateRequest{Name: "security", Description: "Security content"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/tags", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestListAdminTags_Error(t *testing.T) {
	mock := &argustest.MockTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) { return nil, fmt.Errorf("db error") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants/t1/tags", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetAdminTag_Success(t *testing.T) {
	mock := &argustest.MockTags{
		OnGetTag: func(_ context.Context, id string) (*models.Tag, error) {
			return &models.Tag{ID: id, Name: "compliance"}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tags/tg1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestCreateAdminTag_Error(t *testing.T) {
	mock := &argustest.MockTags{
		OnCreateTag: func(_ context.Context, _, _, _ string) (*models.Tag, error) { return nil, fmt.Errorf("failed") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock), argustest.WithAdminVocabulary(&argustest.MockVocabulary{}))
	body := wire.TagCreateRequest{Name: "X", Description: "X desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/tags", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestDeleteAdminTag_Success(t *testing.T) {
	mock := &argustest.MockTags{OnDeleteTag: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/tags/tg1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteAdminTag_NotFound(t *testing.T) {
	mock := &argustest.MockTags{OnDeleteTag: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/tags/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateAdminTag_VocabularyError(t *testing.T) {
	vocab := &argustest.MockVocabulary{
		OnProcess: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("injection detected")
		},
	}
	mock := &argustest.MockTags{}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTags(mock), argustest.WithAdminVocabulary(vocab))
	body := wire.TagCreateRequest{Name: "bad", Description: "bad desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/tags", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status for vocabulary rejection")
	}
}
