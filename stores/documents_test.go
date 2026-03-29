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

type countRow struct {
	Count float64 `db:"count"`
}

func newTestDocuments(t *testing.T, mock *soytesting.MockDB) *Documents {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewDocuments(mock.DB(), astqlpg.New())
}

func TestDocuments_GetDocument(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-1", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
	})

	doc, err := store.GetDocument(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc.ID != "doc-1" {
		t.Errorf("ID: got %q, want %q", doc.ID, "doc-1")
	}
	if doc.Name != "report.pdf" {
		t.Errorf("Name: got %q, want %q", doc.Name, "report.pdf")
	}
	mock.AssertExpectations()
}

func TestDocuments_GetDocument_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetDocument(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocuments_DeleteDocument(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteDocument(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestDocuments_GetDocumentByExternalID(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-1", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
	})

	doc, err := store.GetDocumentByExternalID(context.Background(), "t-1", "ext-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc == nil {
		t.Fatal("expected document, got nil")
	}
	if doc.ExternalID != "ext-1" {
		t.Errorf("ExternalID: got %q, want %q", doc.ExternalID, "ext-1")
	}
	mock.AssertExpectations()
}

func TestDocuments_GetDocumentByExternalID_NotFound(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectQuery().WithRows([]models.Document{})

	doc, err := store.GetDocumentByExternalID(context.Background(), "t-1", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if doc != nil {
		t.Errorf("expected nil, got %+v", doc)
	}
	mock.AssertExpectations()
}

func TestDocuments_GetDocumentByExternalID_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.GetDocumentByExternalID(context.Background(), "t-1", "ext-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocuments_CreateDocument(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-new", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
	})

	doc := &models.Document{
		ID:            "doc-new",
		TenantID:      "t-1",
		ProviderID:    "p-1",
		WatchedPathID: "wp-1",
		ExternalID:    "ext-1",
		Name:          "report.pdf",
		MimeType:      "application/pdf",
		ObjectKey:     "obj-1",
	}
	result, err := store.CreateDocument(context.Background(), doc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "doc-new" {
		t.Errorf("ID: got %q, want %q", result.ID, "doc-new")
	}
	mock.AssertExpectations()
}

func TestDocuments_CreateDocument_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate"))

	doc := &models.Document{
		ID:         "doc-new",
		TenantID:   "t-1",
		ProviderID: "p-1",
		ExternalID: "ext-1",
		Name:       "report.pdf",
		MimeType:   "application/pdf",
		ObjectKey:  "obj-1",
	}
	_, err := store.CreateDocument(context.Background(), doc)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocuments_ListDocuments(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-1", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
		{ID: "doc-2", Name: "spec.docx", MimeType: "application/docx", ExternalID: "ext-2", ObjectKey: "obj-2", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListDocuments(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestDocuments_ListDocuments_CountError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-1", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectQuery().WithError(errors.New("count error"))

	_, err := store.ListDocuments(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error from count query")
	}
	mock.AssertExpectations()
}

func TestDocuments_ListDocuments_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListDocuments(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocuments_ListDocumentsByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-1", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListDocumentsByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestDocuments_ListDocumentsByTenant_CountError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Document{
		{ID: "doc-1", Name: "report.pdf", MimeType: "application/pdf", ExternalID: "ext-1", ObjectKey: "obj-1", TenantID: "t-1", ProviderID: "p-1", WatchedPathID: "wp-1", CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectQuery().WithError(errors.New("count error"))

	_, err := store.ListDocumentsByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error from count query")
	}
	mock.AssertExpectations()
}

func TestDocuments_ListDocumentsByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocuments(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListDocumentsByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
