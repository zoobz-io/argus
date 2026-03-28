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

func TestListTopics_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) {
			return []*models.Topic{{ID: "tp1", Name: "Security"}}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/topics", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.TopicListResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Topics) != 1 || resp.Topics[0].Name != "Security" {
		t.Errorf("unexpected: %+v", resp.Topics)
	}
}

func TestListTopics_Error(t *testing.T) {
	mock := &argustest.MockTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) { return nil, fmt.Errorf("db error") },
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/topics", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}

func TestCreateTopic_Error(t *testing.T) {
	mock := &argustest.MockTopics{
		OnCreateTopic: func(_ context.Context, _, _, _ string) (*models.Topic, error) { return nil, fmt.Errorf("failed") },
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TopicCreateRequest{Name: "X", Description: "X desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/topics", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status")
	}
}

func TestGetTopic_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnGetTopic: func(_ context.Context, id string) (*models.Topic, error) {
			return &models.Topic{ID: id, Name: "Security", Description: "Sec docs"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/topics/tp1", nil)
	rtesting.AssertStatus(t, capture, 200)
}

func TestGetTopic_NotFound(t *testing.T) {
	mock := &argustest.MockTopics{
		OnGetTopic: func(_ context.Context, _ string) (*models.Topic, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithBoundaries(wire.RegisterBoundaries))
	capture := rtesting.ServeRequest(engine, "GET", "/topics/missing", nil)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateTopic_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnCreateTopic: func(_ context.Context, tenantID, name, desc string) (*models.Topic, error) {
			return &models.Topic{ID: "tp-new", TenantID: tenantID, Name: name, Description: desc}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TopicCreateRequest{Name: "Finance", Description: "Finance docs"}
	capture := rtesting.ServeRequest(engine, "POST", "/topics", body)
	rtesting.AssertStatus(t, capture, 201)
}

func TestUpdateTopic_Success(t *testing.T) {
	mock := &argustest.MockTopics{
		OnUpdateTopic: func(_ context.Context, id, name, desc string) (*models.Topic, error) {
			return &models.Topic{ID: id, Name: name, Description: desc}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TopicCreateRequest{Name: "Updated", Description: "Updated desc"}
	capture := rtesting.ServeRequest(engine, "PUT", "/topics/tp1", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestUpdateTopic_NotFound(t *testing.T) {
	mock := &argustest.MockTopics{
		OnUpdateTopic: func(_ context.Context, _, _, _ string) (*models.Topic, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithAPIVocabulary(&argustest.MockVocabulary{}), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TopicCreateRequest{Name: "X", Description: "X desc"}
	capture := rtesting.ServeRequest(engine, "PUT", "/topics/missing", body)
	rtesting.AssertStatus(t, capture, 404)
}

func TestCreateTopic_VocabularyError(t *testing.T) {
	vocab := &argustest.MockVocabulary{
		OnProcess: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("injection detected")
		},
	}
	mock := &argustest.MockTopics{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithAPIVocabulary(vocab), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TopicCreateRequest{Name: "bad", Description: "bad desc"}
	capture := rtesting.ServeRequest(engine, "POST", "/topics", body)
	if capture.StatusCode() == 201 {
		t.Error("expected error status for vocabulary rejection")
	}
}

func TestUpdateTopic_VocabularyError(t *testing.T) {
	vocab := &argustest.MockVocabulary{
		OnProcessUpdate: func(_ context.Context, _, _, _ string) error {
			return fmt.Errorf("injection detected")
		},
	}
	mock := &argustest.MockTopics{}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPITopics(mock), argustest.WithAPIVocabulary(vocab), argustest.WithBoundaries(wire.RegisterBoundaries))
	body := wire.TopicCreateRequest{Name: "bad", Description: "bad desc"}
	capture := rtesting.ServeRequest(engine, "PUT", "/topics/tp1", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status for vocabulary rejection")
	}
}
