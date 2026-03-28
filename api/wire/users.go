package wire

import (
	"context"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/sum"
)

// UserResponse is the public API response for a user.
type UserResponse struct {
	Role        models.UserRole `json:"role" description:"User role" example:"viewer"`
	Email       string          `json:"email" description:"Email address" send.mask:"email"`
	DisplayName string          `json:"display_name" description:"Display name" send.mask:"name"`
	ID          string          `json:"id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OnSend applies boundary masking before the response is marshaled.
func (u *UserResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[UserResponse]](ctx)
	masked, err := b.Send(ctx, *u)
	if err != nil {
		return err
	}
	*u = masked
	return nil
}

// Clone returns a copy of the response.
func (u UserResponse) Clone() UserResponse {
	return u
}

// UserListResponse is the public API response for a paginated user list.
type UserListResponse struct {
	Users  []UserResponse `json:"users" description:"List of users"`
	Offset int            `json:"offset" description:"Number of results skipped"`
	Limit  int            `json:"limit" description:"Page size" example:"20"`
	Total  int64          `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *UserListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[UserListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r UserListResponse) Clone() UserListResponse {
	c := r
	if r.Users != nil {
		c.Users = make([]UserResponse, len(r.Users))
		copy(c.Users, r.Users)
	}
	return c
}
