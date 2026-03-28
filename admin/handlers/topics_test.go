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

func TestListAdminTopics_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) {
			return []*models.Topic{{ID: "tp1", Name: "Security"}}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants/t1/topics", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetAdminTopic_NotFound(t *testing.T) {
	mock := &argustest.MockTopics{
		OnGetTopic: func(_ context.Context, _ string) (*models.Topic, error) { return nil, fmt.Errorf("not found") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/topics/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateAdminTopic_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnCreateTopic: func(_ context.Context, tid, name, desc string) (*models.Topic, error) {
			return &models.Topic{ID: "tp-new", TenantID: tid, Name: name, Description: desc}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock), argustest.WithAdminVocabulary(&argustest.MockVocabulary{}))
	body := wire.TopicCreateRequest{Name: "Finance", Description: "Finance docs"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/topics", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestListAdminTopics_Error(t *testing.T) {
	mock := &argustest.MockTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) { return nil, fmt.Errorf("db error") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/tenants/t1/topics", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestGetAdminTopic_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnGetTopic: func(_ context.Context, id string) (*models.Topic, error) {
			return &models.Topic{ID: id, Name: "Security"}, nil
		},
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock))
	capture := rtesting.ServeRequest(engine, "GET", "/topics/tp1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestCreateAdminTopic_Error(t *testing.T) {
	mock := &argustest.MockTopics{
		OnCreateTopic: func(_ context.Context, _, _, _ string) (*models.Topic, error) { return nil, fmt.Errorf("failed") },
	}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock), argustest.WithAdminVocabulary(&argustest.MockVocabulary{}))
	body := wire.TopicCreateRequest{Name: "X", Description: "X desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/topics", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestDeleteAdminTopic_Success(t *testing.T) {
	mock := &argustest.MockTopics{OnDeleteTopic: func(_ context.Context, _ string) error { return nil }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/topics/tp1", nil)
	rtesting.AssertStatus(t, capture, 204)
}

func TestDeleteAdminTopic_NotFound(t *testing.T) {
	mock := &argustest.MockTopics{OnDeleteTopic: func(_ context.Context, _ string) error { return fmt.Errorf("not found") }}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock))
	capture := rtesting.ServeRequest(engine, "DELETE", "/topics/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateAdminTopic_VocabularyError(t *testing.T) {
	vocab := &argustest.MockVocabulary{
		OnProcess: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("injection detected")
		},
	}
	mock := &argustest.MockTopics{}
	engine := argustest.SetupAdminEngine(t, All(), argustest.WithAdminTopics(mock), argustest.WithAdminVocabulary(vocab))
	body := wire.TopicCreateRequest{Name: "bad", Description: "bad desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/tenants/t1/topics", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status for vocabulary rejection")
	}
}
