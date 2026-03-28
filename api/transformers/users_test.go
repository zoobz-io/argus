//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestUserToResponse(t *testing.T) {
	user := argustest.NewUser()
	resp := UserToResponse(user)

	if resp.ID != "u1" || resp.Email != "user@example.com" || resp.DisplayName != "Jane Doe" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.Role != models.UserRoleViewer {
		t.Errorf("Role = %q, want %q", resp.Role, models.UserRoleViewer)
	}
}

func TestUsersToResponse(t *testing.T) {
	users := []*models.User{argustest.NewUser(), argustest.NewUser()}
	users[1].ID = "u2"

	result := UsersToResponse(users)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].ID != "u1" || result[1].ID != "u2" {
		t.Errorf("IDs mismatch: %q, %q", result[0].ID, result[1].ID)
	}
}

func TestUsersToResponse_Empty(t *testing.T) {
	result := UsersToResponse(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}
