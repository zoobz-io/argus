//go:build testing

package stores

import (
	"context"
	"errors"
	"testing"
	"time"

	astqlpg "github.com/zoobz-io/astql/postgres"
	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/models"

	soytesting "github.com/zoobz-io/soy/testing"
)

func newTestHooks(t *testing.T, mock *soytesting.MockDB) *Hooks {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	svc := sum.New()
	enc, err := cereal.AES([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}
	svc.WithEncryptor(cereal.EncryptAES, enc)
	store := NewHooks(mock.DB(), astqlpg.New())
	k := sum.Start()
	sum.NewBoundary[models.Hook](k)
	sum.Freeze(k)
	return store
}

func encryptSecret(t *testing.T, plaintext string) string {
	t.Helper()
	b := sum.MustUse[sum.Boundary[models.Hook]](context.Background())
	stored, err := b.Store(context.Background(), models.Hook{Secret: plaintext})
	if err != nil {
		t.Fatalf("encrypting via boundary: %v", err)
	}
	return stored.Secret
}

func TestHooks_CreateHook(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "dummy")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-new", TenantID: "t-1", UserID: "u-1", URL: "https://example.com/hook", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	hook, err := store.CreateHook(context.Background(), "t-1", "u-1", "https://example.com/hook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.URL != "https://example.com/hook" {
		t.Errorf("URL: got %q, want %q", hook.URL, "https://example.com/hook")
	}
	if hook.Active != true {
		t.Errorf("Active: got %v, want true", hook.Active)
	}
	mock.AssertExpectations()
}

func TestHooks_CreateHook_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate"))

	_, err := store.CreateHook(context.Background(), "t-1", "u-1", "https://example.com/hook")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestHooks_GetHook(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "secret")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://example.com/hook", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	hook, err := store.GetHook(context.Background(), "h-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.ID != "h-1" {
		t.Errorf("ID: got %q, want %q", hook.ID, "h-1")
	}
	mock.AssertExpectations()
}

func TestHooks_GetHook_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetHook(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestHooks_GetHookByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "secret")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://example.com/hook", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	hook, err := store.GetHookByTenant(context.Background(), "t-1", "h-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.ID != "h-1" {
		t.Errorf("ID: got %q, want %q", hook.ID, "h-1")
	}
	if hook.TenantID != "t-1" {
		t.Errorf("TenantID: got %q, want %q", hook.TenantID, "t-1")
	}
	mock.AssertExpectations()
}

func TestHooks_GetHookByTenant_NotFound(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	mock.ExpectQuery().WithRows([]models.Hook{})

	_, err := store.GetHookByTenant(context.Background(), "t-1", "h-missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestHooks_GetHookByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.GetHookByTenant(context.Background(), "t-1", "h-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestHooks_GetWithSecret(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "whsec_test123")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://example.com/hook", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	hook, err := store.GetWithSecret(context.Background(), "t-1", "h-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if hook.Secret != "whsec_test123" {
		t.Errorf("Secret: got %q, want %q", hook.Secret, "whsec_test123")
	}
	mock.AssertExpectations()
}

func TestHooks_ListHooksByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "secret")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://a.com", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
		{ID: "h-2", TenantID: "t-1", UserID: "u-1", URL: "https://b.com", Secret: secret, Active: true, CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListHooksByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestHooks_ListHooksByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListHooksByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestHooks_ListHooksByTenant_CountError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "secret")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://a.com", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectQuery().WithError(errors.New("count error"))

	_, err := store.ListHooksByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error from count query")
	}
	mock.AssertExpectations()
}

func TestHooks_ListHooks(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "secret")
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://a.com", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 3}})

	result, err := store.ListHooks(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	if result.Total != 3 {
		t.Errorf("Total: got %d, want 3", result.Total)
	}
	mock.AssertExpectations()
}

func TestHooks_DeleteHook(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	secret := encryptSecret(t, "secret")
	// GetHookByTenant (Query)
	mock.ExpectQuery().WithRows([]models.Hook{
		{ID: "h-1", TenantID: "t-1", UserID: "u-1", URL: "https://a.com", Secret: secret, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	// Delete
	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteHook(context.Background(), "t-1", "h-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestHooks_DeleteHook_NotFound(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestHooks(t, mock)

	mock.ExpectQuery().WithRows([]models.Hook{})

	err := store.DeleteHook(context.Background(), "t-1", "h-missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
