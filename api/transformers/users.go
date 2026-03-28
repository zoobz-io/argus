package transformers

import (
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
)

// UserToResponse transforms a User model to an API response.
func UserToResponse(u *models.User) wire.UserResponse {
	return wire.UserResponse{
		ID:          u.ID,
		Email:       u.Email,
		DisplayName: u.DisplayName,
		Role:        u.Role,
	}
}

// UsersToResponse transforms a slice of User models to responses.
func UsersToResponse(users []*models.User) []wire.UserResponse {
	result := make([]wire.UserResponse, len(users))
	for i, u := range users {
		result[i] = UserToResponse(u)
	}
	return result
}
