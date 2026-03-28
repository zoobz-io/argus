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

func newTestTags(t *testing.T, mock *soytesting.MockDB) *Tags {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewTags(mock.DB(), astqlpg.New())
}

func TestTags_GetTag(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Tag{
		{ID: "tag-1", TenantID: "t-1", Name: "Important", Description: "Important docs", CreatedAt: ts, UpdatedAt: ts},
	})

	tag, err := store.GetTag(context.Background(), "tag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("ID: got %q, want %q", tag.ID, "tag-1")
	}
	if tag.Name != "Important" {
		t.Errorf("Name: got %q, want %q", tag.Name, "Important")
	}
	mock.AssertExpectations()
}

func TestTags_CreateTag(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Tag{
		{ID: "tag-new", TenantID: "t-1", Name: "Important", Description: "Important docs", CreatedAt: ts, UpdatedAt: ts},
	})

	tag, err := store.CreateTag(context.Background(), "t-1", "Important", "Important docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.Name != "Important" {
		t.Errorf("Name: got %q, want %q", tag.Name, "Important")
	}
	if tag.Description != "Important docs" {
		t.Errorf("Description: got %q, want %q", tag.Description, "Important docs")
	}
	mock.AssertExpectations()
}

func TestTags_CreateTag_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate name"))

	_, err := store.CreateTag(context.Background(), "t-1", "Important", "Important docs")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		if got := err.Error(); got == "" {
			t.Error("expected non-empty error message")
		}
	}
	mock.AssertExpectations()
}

func TestTags_UpdateTag(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// GetTag (SELECT)
	mock.ExpectQuery().WithRows([]models.Tag{
		{ID: "tag-1", TenantID: "t-1", Name: "Old", Description: "Old desc", CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING)
	mock.ExpectQuery().WithRows([]models.Tag{
		{ID: "tag-1", TenantID: "t-1", Name: "New", Description: "New desc", CreatedAt: ts, UpdatedAt: ts},
	})

	tag, err := store.UpdateTag(context.Background(), "tag-1", "New", "New desc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tag.Name != "New" {
		t.Errorf("Name: got %q, want %q", tag.Name, "New")
	}
	if tag.Description != "New desc" {
		t.Errorf("Description: got %q, want %q", tag.Description, "New desc")
	}
	mock.AssertExpectations()
}

func TestTags_DeleteTag(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteTag(context.Background(), "tag-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestTags_ListTagsByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Tag{
		{ID: "tag-1", TenantID: "t-1", Name: "Important", Description: "Important docs", CreatedAt: ts, UpdatedAt: ts},
		{ID: "tag-2", TenantID: "t-1", Name: "Urgent", Description: "Urgent docs", CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})

	tags, err := store.ListTagsByTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tags) != 2 {
		t.Errorf("Tags: got %d, want 2", len(tags))
	}
	mock.AssertExpectations()
}

func TestTags_UpdateTag_GetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.UpdateTag(context.Background(), "tag-1", "New", "New desc")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestTags_UpdateTag_SetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTags(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// GetTag (SELECT)
	mock.ExpectQuery().WithRows([]models.Tag{
		{ID: "tag-1", TenantID: "t-1", Name: "Old", Description: "Old desc", CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING) — error
	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.UpdateTag(context.Background(), "tag-1", "New", "New desc")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
