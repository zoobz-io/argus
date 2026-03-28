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

func newTestProviders(t *testing.T, mock *soytesting.MockDB) *Providers {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	svc := sum.New()
	enc, err := cereal.AES([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}
	svc.WithEncryptor(cereal.EncryptAES, enc)
	store := NewProviders(mock.DB(), astqlpg.New())
	k := sum.Start()
	sum.NewBoundary[models.Provider](k)
	sum.Freeze(k)
	return store
}

// encryptCreds encrypts credentials using the boundary's Store path.
func encryptCreds(t *testing.T, plaintext string) string {
	t.Helper()
	b := sum.MustUse[sum.Boundary[models.Provider]](context.Background())
	stored, err := b.Store(context.Background(), models.Provider{Credentials: plaintext})
	if err != nil {
		t.Fatalf("encrypting via boundary: %v", err)
	}
	return stored.Credentials
}

func TestProviders_GetProvider(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creds := encryptCreds(t, "secret")
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-1", TenantID: "t-1", Type: models.ProviderGoogleDrive, Name: "My Drive", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	provider, err := store.GetProvider(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.ID != "p-1" {
		t.Errorf("ID: got %q, want %q", provider.ID, "p-1")
	}
	if provider.Name != "My Drive" {
		t.Errorf("Name: got %q, want %q", provider.Name, "My Drive")
	}
	mock.AssertExpectations()
}

func TestProviders_GetProvider_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetProvider(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestProviders_CreateProvider(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creds := encryptCreds(t, "secret")
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-new", TenantID: "t-1", Type: models.ProviderGoogleDrive, Name: "My Drive", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	provider, err := store.CreateProvider(context.Background(), "t-1", models.ProviderGoogleDrive, "My Drive", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.Name != "My Drive" {
		t.Errorf("Name: got %q, want %q", provider.Name, "My Drive")
	}
	if provider.Active != true {
		t.Errorf("Active: got %v, want true", provider.Active)
	}
	mock.AssertExpectations()
}

func TestProviders_CreateProvider_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate name"))

	_, err := store.CreateProvider(context.Background(), "t-1", models.ProviderGoogleDrive, "My Drive", "secret")
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

func TestProviders_UpdateProvider(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creds := encryptCreds(t, "secret")

	// GetProvider (SELECT)
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-1", TenantID: "t-1", Type: models.ProviderGoogleDrive, Name: "Old Drive", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING)
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-1", TenantID: "t-1", Type: models.ProviderOneDrive, Name: "New Drive", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})

	provider, err := store.UpdateProvider(context.Background(), "p-1", models.ProviderOneDrive, "New Drive", "new-secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if provider.Name != "New Drive" {
		t.Errorf("Name: got %q, want %q", provider.Name, "New Drive")
	}
	if provider.Type != models.ProviderOneDrive {
		t.Errorf("Type: got %q, want %q", provider.Type, models.ProviderOneDrive)
	}
	mock.AssertExpectations()
}

func TestProviders_UpdateProvider_GetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.UpdateProvider(context.Background(), "p-1", models.ProviderOneDrive, "New Drive", "new-secret")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestProviders_UpdateProvider_SetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creds := encryptCreds(t, "secret")
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-1", TenantID: "t-1", Type: models.ProviderGoogleDrive, Name: "Old Drive", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.UpdateProvider(context.Background(), "p-1", models.ProviderOneDrive, "New Drive", "new-secret")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestProviders_DeleteProvider(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteProvider(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestProviders_ListProviders(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creds := encryptCreds(t, "secret")
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-1", TenantID: "t-1", Type: models.ProviderGoogleDrive, Name: "Drive A", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
		{ID: "p-2", TenantID: "t-1", Type: models.ProviderOneDrive, Name: "Drive B", Credentials: creds, Active: true, CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListProviders(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestProviders_ListProvidersByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	creds := encryptCreds(t, "secret")
	mock.ExpectQuery().WithRows([]models.Provider{
		{ID: "p-1", TenantID: "t-1", Type: models.ProviderGoogleDrive, Name: "Drive A", Credentials: creds, Active: true, CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectQuery().WithRows([]countRow{{Count: 5}})

	result, err := store.ListProvidersByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestProviders_ListProviders_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListProviders(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestProviders_ListProvidersByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestProviders(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListProvidersByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
