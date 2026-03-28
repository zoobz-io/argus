//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"

	soytesting "github.com/zoobz-io/soy/testing"
)

func newTestTopics(t *testing.T, mock *soytesting.MockDB) *Topics {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewTopics(mock.DB(), astqlpg.New())
}

func TestTopics_GetTopic(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Topic{
		{ID: "topic-1", TenantID: "t-1", Name: "Finance", Description: "Financial docs", CreatedAt: ts, UpdatedAt: ts},
	})

	topic, err := store.GetTopic(context.Background(), "topic-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic.ID != "topic-1" {
		t.Errorf("ID: got %q, want %q", topic.ID, "topic-1")
	}
	if topic.Name != "Finance" {
		t.Errorf("Name: got %q, want %q", topic.Name, "Finance")
	}
	mock.AssertExpectations()
}

func TestTopics_CreateTopic(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Topic{
		{ID: "topic-new", TenantID: "t-1", Name: "Finance", Description: "Financial docs", CreatedAt: ts, UpdatedAt: ts},
	})

	topic, err := store.CreateTopic(context.Background(), "t-1", "Finance", "Financial docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic.Name != "Finance" {
		t.Errorf("Name: got %q, want %q", topic.Name, "Finance")
	}
	if topic.Description != "Financial docs" {
		t.Errorf("Description: got %q, want %q", topic.Description, "Financial docs")
	}
	mock.AssertExpectations()
}

func TestTopics_UpdateTopic(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// GetTopic (SELECT)
	mock.ExpectQuery().WithRows([]models.Topic{
		{ID: "topic-1", TenantID: "t-1", Name: "Old", Description: "Old desc", CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING)
	mock.ExpectQuery().WithRows([]models.Topic{
		{ID: "topic-1", TenantID: "t-1", Name: "New", Description: "New desc", CreatedAt: ts, UpdatedAt: ts},
	})

	topic, err := store.UpdateTopic(context.Background(), "topic-1", "New", "New desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if topic.Name != "New" {
		t.Errorf("Name: got %q, want %q", topic.Name, "New")
	}
	if topic.Description != "New desc" {
		t.Errorf("Description: got %q, want %q", topic.Description, "New desc")
	}
	mock.AssertExpectations()
}

func TestTopics_DeleteTopic(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteTopic(context.Background(), "topic-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestTopics_ListTopicsByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Topic{
		{ID: "topic-1", TenantID: "t-1", Name: "Finance", Description: "Financial docs", CreatedAt: ts, UpdatedAt: ts},
		{ID: "topic-2", TenantID: "t-1", Name: "Legal", Description: "Legal docs", CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})

	topics, err := store.ListTopicsByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topics) != 2 {
		t.Errorf("Topics: got %d, want 2", len(topics))
	}
	mock.AssertExpectations()
}

func TestTopics_CreateTopic_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.CreateTopic(context.Background(), "t-1", "Finance", "Financial docs")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestTopics_UpdateTopic_GetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.UpdateTopic(context.Background(), "topic-1", "New", "New desc")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestTopics_UpdateTopic_SetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTopics(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// GetTopic (SELECT)
	mock.ExpectQuery().WithRows([]models.Topic{
		{ID: "topic-1", TenantID: "t-1", Name: "Old", Description: "Old desc", CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING) — error
	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.UpdateTopic(context.Background(), "topic-1", "New", "New desc")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
