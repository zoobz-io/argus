package wire

import (
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestUserResponse_Clone(t *testing.T) {
	orig := UserResponse{ID: "u1", Email: "a@b.com", Role: models.UserRoleViewer}
	cloned := orig.Clone()
	if cloned.ID != orig.ID || cloned.Email != orig.Email {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
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
