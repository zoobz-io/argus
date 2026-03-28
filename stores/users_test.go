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

func newTestUsers(t *testing.T, mock *soytesting.MockDB) *Users {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	sum.New()
	return NewUsers(mock.DB(), astqlpg.New())
}

func testUser() models.User {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	return models.User{
		ID:          "u-1",
		ExternalID:  "ext-1",
		TenantID:    "t-1",
		Email:       "alice@example.com",
		DisplayName: "Alice",
		Role:        models.UserRoleViewer,
		Status:      models.UserStatusActive,
		CreatedAt:   ts,
		UpdatedAt:   ts,
	}
}

func TestUsers_GetUser(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	mock.ExpectQuery().WithRows([]models.User{u})

	user, err := store.GetUser(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u-1" {
		t.Errorf("ID: got %q, want %q", user.ID, "u-1")
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email: got %q, want %q", user.Email, "alice@example.com")
	}
	mock.AssertExpectations()
}

func TestUsers_GetUser_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetUser(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_GetUserByExternalID(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	mock.ExpectQuery().WithRows([]models.User{u})

	user, err := store.GetUserByExternalID(context.Background(), "ext-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ExternalID != "ext-1" {
		t.Errorf("ExternalID: got %q, want %q", user.ExternalID, "ext-1")
	}
	mock.AssertExpectations()
}

func TestUsers_GetUserByExternalID_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.GetUserByExternalID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_GetUserByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	mock.ExpectQuery().WithRows([]models.User{u})

	user, err := store.GetUserByTenant(context.Background(), "u-1", "t-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ID != "u-1" {
		t.Errorf("ID: got %q, want %q", user.ID, "u-1")
	}
	if user.TenantID != "t-1" {
		t.Errorf("TenantID: got %q, want %q", user.TenantID, "t-1")
	}
	mock.AssertExpectations()
}

func TestUsers_GetUserByTenant_NotFound(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithRows([]models.User{})

	_, err := store.GetUserByTenant(context.Background(), "u-1", "wrong-tenant")
	if err == nil {
		t.Fatal("expected error for tenant mismatch")
	}
	mock.AssertExpectations()
}

func TestUsers_GetUserByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.GetUserByTenant(context.Background(), "u-1", "t-1")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_ListUsersByTenant(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u1 := testUser()
	u2 := testUser()
	u2.ID = "u-2"
	u2.ExternalID = "ext-2"
	u2.Email = "bob@example.com"
	u2.DisplayName = "Bob"
	mock.ExpectQuery().WithRows([]models.User{u1, u2})

	result, err := store.ListUsersByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Errorf("Items: got %d, want 2", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestUsers_ListUsersByTenant_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListUsersByTenant(context.Background(), "t-1", models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_ListUsers(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	mock.ExpectQuery().WithRows([]models.User{u})

	result, err := store.ListUsers(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("Items: got %d, want 1", len(result.Items))
	}
	mock.AssertExpectations()
}

func TestUsers_ListUsers_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.ListUsers(context.Background(), models.OffsetPage{Offset: 0, Limit: 10})
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_CreateUser(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	// Set uses INSERT...RETURNING, which is a query that returns rows.
	mock.ExpectQuery().WithRows([]models.User{u})

	user, err := store.CreateUser(context.Background(), "ext-1", "t-1", "alice@example.com", "Alice", models.UserRoleViewer)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("Email: got %q, want %q", user.Email, "alice@example.com")
	}
	if user.Role != models.UserRoleViewer {
		t.Errorf("Role: got %q, want %q", user.Role, models.UserRoleViewer)
	}
	mock.AssertExpectations()
}

func TestUsers_CreateUser_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectExec().WithError(errors.New("duplicate external_id"))

	_, err := store.CreateUser(context.Background(), "ext-1", "t-1", "alice@example.com", "Alice", models.UserRoleViewer)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_UpdateUser(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()

	// GetUser (SELECT)
	mock.ExpectQuery().WithRows([]models.User{u})
	// Set (INSERT...RETURNING)
	updated := u
	updated.Email = "newalice@example.com"
	updated.DisplayName = "Alice New"
	updated.Role = models.UserRoleAdmin
	mock.ExpectQuery().WithRows([]models.User{updated})

	user, err := store.UpdateUser(context.Background(), "u-1", "newalice@example.com", "Alice New", models.UserRoleAdmin, models.UserStatusActive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "newalice@example.com" {
		t.Errorf("Email: got %q, want %q", user.Email, "newalice@example.com")
	}
	if user.Role != models.UserRoleAdmin {
		t.Errorf("Role: got %q, want %q", user.Role, models.UserRoleAdmin)
	}
	mock.AssertExpectations()
}

func TestUsers_UpdateUser_GetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("not found"))

	_, err := store.UpdateUser(context.Background(), "u-1", "alice@example.com", "Alice", models.UserRoleViewer, models.UserStatusActive)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_UpdateUser_SetError(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	mock.ExpectQuery().WithRows([]models.User{u})
	mock.ExpectExec().WithError(errors.New("constraint violation"))

	_, err := store.UpdateUser(context.Background(), "u-1", "alice@example.com", "Alice", models.UserRoleViewer, models.UserStatusActive)
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}

func TestUsers_DeleteUser(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectExec().WithResult(1, 0)

	err := store.DeleteUser(context.Background(), "u-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	mock.AssertExpectations()
}

func TestUsers_UpsertFromClaims(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	u := testUser()
	// Insert().OnConflict().DoUpdate().Exec() returns a query with rows.
	mock.ExpectQuery().WithRows([]models.User{u})

	user, err := store.UpsertFromClaims(context.Background(), "ext-1", "t-1", "alice@example.com", "Alice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.ExternalID != "ext-1" {
		t.Errorf("ExternalID: got %q, want %q", user.ExternalID, "ext-1")
	}
	mock.AssertExpectations()
}

func TestUsers_UpsertFromClaims_Error(t *testing.T) {
	mock := soytesting.NewMockDB(t)
	store := newTestUsers(t, mock)

	mock.ExpectQuery().WithError(errors.New("db error"))

	_, err := store.UpsertFromClaims(context.Background(), "ext-1", "t-1", "alice@example.com", "Alice")
	if err == nil {
		t.Fatal("expected error")
	}
	mock.AssertExpectations()
}
