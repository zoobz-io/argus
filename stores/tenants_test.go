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

func newTestTenants(t *testing.T, mock *soytesting.MockDB) *Tenants {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewTenants(mock.DB(), astqlpg.New())
}

func TestTenants_GetTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Tenant{
		{ID: "t-1", Name: "Acme", Slug: "acme", CreatedAt: ts, UpdatedAt: ts},
	})

	tenant, err := store.GetTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.ID != "t-1" {
		t.Errorf("ID: got %q, want %q", tenant.ID, "t-1")
	}
	if tenant.Name != "Acme" {
		t.Errorf("Name: got %q, want %q", tenant.Name, "Acme")
	}
	mock.AssertExpectations()
}

func TestTenants_GetTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetTenant(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestTenants_CreateTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Set uses INSERT...RETURNING, which is a query that returns rows.
	mock.ExpectQuery().WithRows([]models.Tenant{
		{ID: "t-new", Name: "Acme", Slug: "acme", CreatedAt: ts, UpdatedAt: ts},
	})

	tenant, err := store.CreateTenant(context.Background(), "Acme", "acme")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.Name != "Acme" {
		t.Errorf("Name: got %q, want %q", tenant.Name, "Acme")
	}
	if tenant.Slug != "acme" {
		t.Errorf("Slug: got %q, want %q", tenant.Slug, "acme")
	}
	mock.AssertExpectations()
}

func TestTenants_CreateTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate slug"))

	_, err := store.CreateTenant(context.Background(), "Acme", "acme")
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, errors.Unwrap(err)) {
		// Just verify it wraps.
		if got := err.Error(); got == "" {
			t.Error("expected non-empty error message")
		}
	}
	mock.AssertExpectations()
}

func TestTenants_UpdateTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)

	// GetTenant (SELECT)
	mock.ExpectQuery().WithRows([]models.Tenant{
		{ID: "t-1", Name: "Old", Slug: "old", CreatedAt: ts, UpdatedAt: ts},
	})
	// Set (INSERT...RETURNING)
	mock.ExpectQuery().WithRows([]models.Tenant{
		{ID: "t-1", Name: "New", Slug: "new", CreatedAt: ts, UpdatedAt: ts},
	})

	tenant, err := store.UpdateTenant(context.Background(), "t-1", "New", "new")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tenant.Name != "New" {
		t.Errorf("Name: got %q, want %q", tenant.Name, "New")
	}
	if tenant.Slug != "new" {
		t.Errorf("Slug: got %q, want %q", tenant.Slug, "new")
	}
	mock.AssertExpectations()
}

func TestTenants_UpdateTenant_GetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.UpdateTenant(context.Background(), "t-1", "New", "new")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestTenants_UpdateTenant_SetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Tenant{
		{ID: "t-1", Name: "Old", Slug: "old", CreatedAt: ts, UpdatedAt: ts},
	})
	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.UpdateTenant(context.Background(), "t-1", "New", "new")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestTenants_DeleteTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteTenant(context.Background(), "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestTenants_ListTenants(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	mock.ExpectQuery().WithRows([]models.Tenant{
		{ID: "t-1", Name: "Acme", Slug: "acme", CreatedAt: ts, UpdatedAt: ts},
		{ID: "t-2", Name: "Beta", Slug: "beta", CreatedAt: ts.Add(time.Hour), UpdatedAt: ts.Add(time.Hour)},
	})

	result, err := store.ListTenants(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestTenants_ListTenants_QueryError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestTenants(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListTenants(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
