//go:build testing

package wire

import (
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestUserResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := UserResponse{ID: "u1", Email: "a@b.com", DisplayName: "Jane", Role: models.UserRoleViewer}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// ID and Role are unmasked, Email and DisplayName have send.mask tags.
	if resp.ID != "u1" || resp.Role != models.UserRoleViewer {
		t.Errorf("unmasked fields mutated: %+v", resp)
	}
}

func TestUserResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := UserResponse{ID: "u1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestUserResponse_Clone(t *testing.T) {
	orig := UserResponse{ID: "u1", Email: "a@b.com", Role: models.UserRoleViewer}
	cloned := orig.Clone()
	if cloned.ID != orig.ID || cloned.Email != orig.Email {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestUserListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := UserListResponse{
		Users: []UserResponse{{ID: "u1"}},
		Limit: 20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUserListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := UserListResponse{Users: []UserResponse{{ID: "u1"}}}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestUserListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := UserListResponse{
		Users: []UserResponse{{ID: "u1", Email: "a@b.com"}},
		Limit: 20,
	}
	cloned := orig.Clone()

	cloned.Users[0].Email = "mutated"
	if orig.Users[0].Email == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestUserListResponse_Clone_NilUsers(t *testing.T) {
	orig := UserListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Users != nil {
		t.Error("expected nil Users")
	}
}
