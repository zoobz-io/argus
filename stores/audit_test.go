//go:build testing

package stores

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"
)

func newTestAudit(t *testing.T, mock *mockSearchProvider) *Audit {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewAudit(mock)
}

func TestAudit_Index(t *testing.T) {
	var indexedID string
	mock := &mockSearchProvider{
		OnIndex: func(_ context.Context, _, id string, _ []byte) error {
			indexedID = id
			return nil
		},
	}

	store := newTestAudit(t, mock)
	entry := &models.AuditEntry{
		ID:       "a-1",
		Action:   "provider.created",
		TenantID: "t-1",
	}

	if err := store.Index(context.Background(), entry); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if indexedID != "a-1" {
		t.Errorf("indexed ID: got %q, want %q", indexedID, "a-1")
	}
}

func TestAudit_Index_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnIndex: func(_ context.Context, _, _ string, _ []byte) error {
			return errors.New("index failure")
		},
	}

	store := newTestAudit(t, mock)
	err := store.Index(context.Background(), &models.AuditEntry{ID: "a-1"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAudit_Search_NoFilters(t *testing.T) {
	ts := time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Total: 1,
				Hits: []grub.SearchHit{
					{Source: jsonBytes(t, models.AuditEntry{
						ID: "a-1", Action: "tenant.created", TenantID: "t-1", Timestamp: ts,
					})},
				},
			}, nil
		},
	}

	store := newTestAudit(t, mock)
	result, err := store.Search(context.Background(), models.AuditSearchParams{Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Total != 1 {
		t.Errorf("total: got %d, want 1", result.Total)
	}
	if len(result.Items) != 1 {
		t.Fatalf("items: got %d, want 1", len(result.Items))
	}
	if result.Items[0].ID != "a-1" {
		t.Errorf("ID: got %q", result.Items[0].ID)
	}
}

func TestAudit_Search_AllFilters(t *testing.T) {
	from := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 3, 31, 0, 0, 0, 0, time.UTC)
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{Total: 0, Hits: []grub.SearchHit{}}, nil
		},
	}

	store := newTestAudit(t, mock)
	_, err := store.Search(context.Background(), models.AuditSearchParams{
		TenantID:     "t-1",
		Action:       "provider.created",
		ResourceType: "provider",
		ActorID:      "u-1",
		From:         &from,
		To:           &to,
		Limit:        50,
		Offset:       10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAudit_Search_Error(t *testing.T) {
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return nil, errors.New("search failure")
		},
	}

	store := newTestAudit(t, mock)
	_, err := store.Search(context.Background(), models.AuditSearchParams{Limit: 20})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAudit_Search_WithMetadata(t *testing.T) {
	mock := &mockSearchProvider{
		OnSearch: func(_ context.Context, _ string, _ *lucene.Search) (*grub.SearchResponse, error) {
			return &grub.SearchResponse{
				Total: 1,
				Hits: []grub.SearchHit{
					{Source: jsonBytes(t, models.AuditEntry{
						ID:       "a-1",
						Action:   "provider.created",
						Metadata: json.RawMessage(`{"provider_type":"google_drive"}`),
					})},
				},
			}, nil
		},
	}

	store := newTestAudit(t, mock)
	result, err := store.Search(context.Background(), models.AuditSearchParams{Limit: 20})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Items[0].Metadata == nil {
		t.Error("expected metadata")
	}
}
