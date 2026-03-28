package models

import (
	"testing"
	"time"
)

func TestUser_GetID(t *testing.T) {
	u := User{ID: "u-1"}
	if got := u.GetID(); got != "u-1" {
		t.Errorf("GetID() = %q, want %q", got, "u-1")
	}
}

func TestUser_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	u := User{CreatedAt: ts}
	if got := u.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestUser_Clone(t *testing.T) {
	seen := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	u := User{
		ID: "u-1", ExternalID: "ext-1", TenantID: "t-1",
		Email: "a@b.com", DisplayName: "Jane",
		Role: UserRoleViewer, Status: UserStatusActive,
		LastSeenAt: &seen,
	}
	clone := u.Clone()

	if clone.ID != u.ID || clone.Email != u.Email {
		t.Error("Clone did not copy all fields")
	}

	// Mutating the pointer field in clone should not affect original.
	newTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	*clone.LastSeenAt = newTime
	if u.LastSeenAt.Equal(newTime) {
		t.Error("mutating clone LastSeenAt leaked to original")
	}
}

func TestUser_Clone_NilLastSeen(t *testing.T) {
	u := User{ID: "u-2"}
	clone := u.Clone()
	if clone.LastSeenAt != nil {
		t.Error("expected nil LastSeenAt on clone")
	}
}
