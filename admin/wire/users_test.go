package wire

import (
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
)

func TestAdminUserResponse_Clone(t *testing.T) {
	seen := time.Now()
	orig := AdminUserResponse{ID: "u1", Email: "a@b.com", LastSeenAt: &seen}
	cloned := orig.Clone()

	if cloned.ID != orig.ID || cloned.Email != orig.Email {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}

	// Mutating pointer field should not affect original.
	newTime := time.Date(2099, 1, 1, 0, 0, 0, 0, time.UTC)
	*cloned.LastSeenAt = newTime
	if orig.LastSeenAt.Equal(newTime) {
		t.Error("mutating clone LastSeenAt affected original")
	}
}

func TestAdminUserResponse_Clone_NilLastSeen(t *testing.T) {
	orig := AdminUserResponse{ID: "u1"}
	cloned := orig.Clone()
	if cloned.LastSeenAt != nil {
		t.Error("expected nil LastSeenAt")
	}
}

func TestAdminUserListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := AdminUserListResponse{
		Users: []AdminUserResponse{{ID: "u1", Email: "a@b.com"}},
		Limit: 20,
	}
	cloned := orig.Clone()

	cloned.Users[0].Email = "mutated"
	if orig.Users[0].Email == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestAdminUserListResponse_Clone_NilUsers(t *testing.T) {
	orig := AdminUserListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Users != nil {
		t.Error("expected nil Users")
	}
}

func TestAdminUserCreateRequest_Validate_Valid(t *testing.T) {
	r := &AdminUserCreateRequest{
		TenantID:    "t1",
		ExternalID:  "ext-1",
		Email:       "user@example.com",
		DisplayName: "Jane Doe",
		Role:        models.UserRoleViewer,
	}
	if err := r.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestAdminUserCreateRequest_Validate_MissingFields(t *testing.T) {
	r := &AdminUserCreateRequest{}
	if err := r.Validate(); err == nil {
		t.Error("expected error for missing fields")
	}
}

func TestAdminUserCreateRequest_Clone(t *testing.T) {
	r := AdminUserCreateRequest{TenantID: "t1", Email: "a@b.com"}
	cloned := r.Clone()
	if cloned.TenantID != r.TenantID {
		t.Error("clone mismatch")
	}
}

func TestAdminUserUpdateRequest_Validate_Valid(t *testing.T) {
	email := "new@example.com"
	r := &AdminUserUpdateRequest{Email: &email}
	if err := r.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestAdminUserUpdateRequest_Validate_Empty(t *testing.T) {
	r := &AdminUserUpdateRequest{}
	if err := r.Validate(); err != nil {
		t.Errorf("empty update should be valid, got: %v", err)
	}
}

func TestAdminUserUpdateRequest_Clone(t *testing.T) {
	role := models.UserRoleAdmin
	status := models.UserStatusInactive
	email := "a@b.com"
	name := "Jane"
	orig := AdminUserUpdateRequest{Role: &role, Status: &status, Email: &email, DisplayName: &name}
	cloned := orig.Clone()

	// Mutating clone should not affect original.
	newRole := models.UserRoleOwner
	*cloned.Role = newRole
	if *orig.Role == newRole {
		t.Error("mutating clone Role affected original")
	}

	newEmail := "mutated"
	*cloned.Email = newEmail
	if *orig.Email == newEmail {
		t.Error("mutating clone Email affected original")
	}
}

func TestAdminUserUpdateRequest_Clone_NilFields(t *testing.T) {
	orig := AdminUserUpdateRequest{}
	cloned := orig.Clone()
	if cloned.Role != nil || cloned.Status != nil || cloned.Email != nil || cloned.DisplayName != nil {
		t.Error("expected all nil fields on clone")
	}
}
