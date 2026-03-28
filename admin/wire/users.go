package wire

import (
	"time"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/check"
)

// AdminUserResponse is the admin API response for user data.
type AdminUserResponse struct {
	CreatedAt  time.Time        `json:"created_at" description:"Creation timestamp"`
	UpdatedAt  time.Time        `json:"updated_at" description:"Last update timestamp"`
	LastSeenAt *time.Time       `json:"last_seen_at,omitempty" description:"Last seen timestamp"`
	Role       models.UserRole  `json:"role" description:"User role" example:"viewer"`
	Status     models.UserStatus `json:"status" description:"Account status" example:"active"`
	ID         string           `json:"id" description:"User ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	ExternalID string           `json:"external_id" description:"External identity provider ID"`
	TenantID   string           `json:"tenant_id" description:"Tenant ID"`
	Email      string           `json:"email" description:"Email address" example:"user@example.com"`
	DisplayName string          `json:"display_name" description:"Display name" example:"Jane Doe"`
}

// Clone returns a deep copy of the response.
func (u AdminUserResponse) Clone() AdminUserResponse {
	c := u
	if u.LastSeenAt != nil {
		t := *u.LastSeenAt
		c.LastSeenAt = &t
	}
	return c
}

// AdminUserListResponse is the admin API response for a paginated user list.
type AdminUserListResponse struct {
	Users  []AdminUserResponse `json:"users" description:"List of users"`
	Offset int                 `json:"offset" description:"Number of results skipped"`
	Limit  int                 `json:"limit" description:"Page size" example:"20"`
	Total  int64               `json:"total" description:"Total number of results"`
}

// Clone returns a deep copy of the response.
func (r AdminUserListResponse) Clone() AdminUserListResponse {
	c := r
	if r.Users != nil {
		c.Users = make([]AdminUserResponse, len(r.Users))
		for i, u := range r.Users {
			c.Users[i] = u.Clone()
		}
	}
	return c
}

// AdminUserCreateRequest is the request body for creating a user.
type AdminUserCreateRequest struct {
	Role        models.UserRole `json:"role" description:"User role" example:"viewer"`
	TenantID    string          `json:"tenant_id" description:"Tenant ID"`
	ExternalID  string          `json:"external_id" description:"External identity provider ID"`
	Email       string          `json:"email" description:"Email address" example:"user@example.com"`
	DisplayName string          `json:"display_name" description:"Display name" example:"Jane Doe"`
}

// Validate checks that all required fields are present.
func (r *AdminUserCreateRequest) Validate() error {
	return check.All(
		check.Str(r.TenantID, "tenant_id").Required().V(),
		check.Str(r.ExternalID, "external_id").Required().V(),
		check.Str(r.Email, "email").Required().MaxLen(255).V(),
		check.Str(r.DisplayName, "display_name").Required().MaxLen(255).V(),
		check.Str(string(r.Role), "role").Required().V(),
	).Err()
}

// Clone returns a copy of the request.
func (r AdminUserCreateRequest) Clone() AdminUserCreateRequest {
	return r
}

// AdminUserUpdateRequest is the request body for updating a user.
type AdminUserUpdateRequest struct {
	Role        *models.UserRole   `json:"role,omitempty" description:"User role" example:"editor"`
	Status      *models.UserStatus `json:"status,omitempty" description:"Account status" example:"active"`
	Email       *string            `json:"email,omitempty" description:"Email address" example:"user@example.com"`
	DisplayName *string            `json:"display_name,omitempty" description:"Display name" example:"Jane Doe"`
}

// Validate checks that update fields are valid.
func (r *AdminUserUpdateRequest) Validate() error {
	return check.All(
		check.OptStr(r.Email, "email").MaxLen(255).V(),
		check.OptStr(r.DisplayName, "display_name").MaxLen(255).V(),
	).Err()
}

// Clone returns a deep copy of the request.
func (r AdminUserUpdateRequest) Clone() AdminUserUpdateRequest {
	c := r
	if r.Role != nil {
		v := *r.Role
		c.Role = &v
	}
	if r.Status != nil {
		v := *r.Status
		c.Status = &v
	}
	if r.Email != nil {
		v := *r.Email
		c.Email = &v
	}
	if r.DisplayName != nil {
		v := *r.DisplayName
		c.DisplayName = &v
	}
	return c
}
