//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestUserToAdminResponse(t *testing.T) {
	user := argustest.NewUser()
	resp := UserToAdminResponse(user)

	if resp.ID != "u1" || resp.Email != "user@example.com" || resp.DisplayName != "Jane Doe" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.ExternalID != "ext-u1" || resp.TenantID != "t1" {
		t.Errorf("external/tenant mismatch: %+v", resp)
	}
	if resp.Role != models.UserRoleViewer || resp.Status != models.UserStatusActive {
		t.Errorf("role/status mismatch: %+v", resp)
	}
	if !resp.CreatedAt.Equal(argustest.FixtureTime) || !resp.UpdatedAt.Equal(argustest.FixtureTime2) {
		t.Error("timestamps not mapped")
	}
	if resp.LastSeenAt == nil || !resp.LastSeenAt.Equal(argustest.FixtureTime2) {
		t.Error("LastSeenAt not mapped")
	}
}

func TestUsersToAdminList(t *testing.T) {
	result := &models.OffsetResult[models.User]{
		Items:  []*models.User{argustest.NewUser()},
		Offset: 0,
		Total:  1,
	}

	resp := UsersToAdminList(result)
	if len(resp.Users) != 1 || resp.Users[0].ID != "u1" {
		t.Errorf("users mismatch: %+v", resp.Users)
	}
	if resp.Total != 1 || resp.Offset != 0 {
		t.Errorf("pagination mismatch: %+v", resp)
	}
}
