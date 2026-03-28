package transformers

import (
	"github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/models"
)

// UserToAdminResponse converts a user model to an admin API response.
func UserToAdminResponse(u *models.User) wire.AdminUserResponse {
	return wire.AdminUserResponse{
		ID:          u.ID,
		ExternalID:  u.ExternalID,
		TenantID:    u.TenantID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
		Status:      u.Status,
		LastSeenAt:  u.LastSeenAt,
		CreatedAt:   u.CreatedAt,
		UpdatedAt:   u.UpdatedAt,
	}
}

// UsersToAdminList converts an offset result of users to an admin list response.
func UsersToAdminList(result *models.OffsetResult[models.User]) wire.AdminUserListResponse {
	items := make([]wire.AdminUserResponse, len(result.Items))
	for i, u := range result.Items {
		items[i] = UserToAdminResponse(u)
	}
	return wire.AdminUserListResponse{
		Users:  items,
		Offset: result.Offset,
		Total:  result.Total,
	}
}
