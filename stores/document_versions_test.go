//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"

	soytesting "github.com/zoobz-io/soy/testing"
)

type mockBucketProvider struct {
	OnGet func(ctx context.Context, key string) ([]byte, *grub.ObjectInfo, error)
}

func (m *mockBucketProvider) Get(ctx context.Context, key string) ([]byte, *grub.ObjectInfo, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, key)
	}
	return nil, nil, nil
}
func (m *mockBucketProvider) Put(_ context.Context, _ string, _ []byte, _ *grub.ObjectInfo) error {
	return nil
}
func (m *mockBucketProvider) Delete(_ context.Context, _ string) error { return nil }
func (m *mockBucketProvider) Exists(_ context.Context, _ string) (bool, error) {
	return false, nil
}
func (m *mockBucketProvider) List(_ context.Context, _ string, _ int) ([]grub.ObjectInfo, error) {
	return nil, nil
}

func newTestDocumentVersions(t *testing.T, mock *soytesting.MockDB, bucket grub.BucketProvider) *DocumentVersions {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewDocumentVersions(mock.DB(), astqlpg.New(), bucket)
}

func TestDocumentVersions_GetDocumentVersion(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "abc123", VersionNumber: 1, CreatedAt: ts},
	})

	ver, err := store.GetDocumentVersion(context.Background(), "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver.ID != "ver-1" {
		t.Errorf("ID: got %q, want %q", ver.ID, "ver-1")
	}
	if ver.DocumentID != "doc-1" {
		t.Errorf("DocumentID: got %q, want %q", ver.DocumentID, "doc-1")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_GetDocumentVersion_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetDocumentVersion(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_DeleteDocumentVersion(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteDocumentVersion(context.Background(), "ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_ListDocumentVersions(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "abc123", VersionNumber: 1, CreatedAt: ts},
		{ID: "ver-2", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "def456", VersionNumber: 2, CreatedAt: ts.Add(time.Hour)},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListDocumentVersions(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_ListVersionsByDocument(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "abc123", VersionNumber: 1, CreatedAt: ts},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListVersionsByDocument(context.Background(), "doc-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_CreateDocumentVersion(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-new", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "abc123", VersionNumber: 1, CreatedAt: ts},
	})

	ver := &models.DocumentVersion{
		ID:            "ver-new",
		DocumentID:    "doc-1",
		TenantID:      "t-1",
		ContentHash:   "abc123",
		VersionNumber: 1,
	}
	result, err := store.CreateDocumentVersion(context.Background(), ver)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "ver-new" {
		t.Errorf("ID: got %q, want %q", result.ID, "ver-new")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_CreateDocumentVersion_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectExec().WithError(errors.New("constraint violation"))

	ver := &models.DocumentVersion{
		ID:            "ver-new",
		DocumentID:    "doc-1",
		TenantID:      "t-1",
		ContentHash:   "abc123",
		VersionNumber: 1,
	}
	_, err := store.CreateDocumentVersion(context.Background(), ver)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_GetVersionContent(t *testing.T) {
	bucket := &mockBucketProvider{
		OnGet: func(_ context.Context, key string) ([]byte, *grub.ObjectInfo, error) {
			if key != "objects/doc-1/ver-1" {
				t.Errorf("key: got %q, want %q", key, "objects/doc-1/ver-1")
			}
			return []byte("file content"), nil, nil
		},
	}
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, bucket)

	data, err := store.GetVersionContent(context.Background(), "objects/doc-1/ver-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != "file content" {
		t.Errorf("data: got %q, want %q", string(data), "file content")
	}
}

func TestDocumentVersions_GetVersionContent_Error(t *testing.T) {
	bucket := &mockBucketProvider{
		OnGet: func(_ context.Context, _ string) ([]byte, *grub.ObjectInfo, error) {
			return nil, nil, errors.New("storage unavailable")
		},
	}
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, bucket)

	_, err := store.GetVersionContent(context.Background(), "objects/doc-1/ver-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDocumentVersions_ListDocumentVersions_CountError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "abc123", VersionNumber: 1, CreatedAt: ts},
	})
	mock.ExpectQuery().WithError(errors.New("count error"))

	_, err := store.ListDocumentVersions(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error from count query")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_ListDocumentVersions_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListDocumentVersions(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_ListVersionsByDocument_CountError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-1", DocumentID: "doc-1", TenantID: "t-1", ContentHash: "abc123", VersionNumber: 1, CreatedAt: ts},
	})
	mock.ExpectQuery().WithError(errors.New("count error"))

	_, err := store.ListVersionsByDocument(context.Background(), "doc-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error from count query")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_ListVersionsByDocument_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListVersionsByDocument(context.Background(), "doc-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_GetLatestVersion(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.DocumentVersion{
		{ID: "ver-3", DocumentID: "doc-1", TenantID: "t-1", VersionNumber: 3, ContentHash: "hash-3", CreatedAt: ts},
	})

	ver, err := store.GetLatestVersion(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver == nil {
		t.Fatal("expected version")
	}
	if ver.ID != "ver-3" {
		t.Errorf("ID: got %q, want %q", ver.ID, "ver-3")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_GetLatestVersion_None(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectQuery().WithRows([]models.DocumentVersion{})

	ver, err := store.GetLatestVersion(context.Background(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ver != nil {
		t.Error("expected nil for no versions")
	}
	mock.AssertExpectations()
}

func TestDocumentVersions_GetLatestVersion_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestDocumentVersions(t, mock, nil)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.GetLatestVersion(context.Background(), "doc-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
