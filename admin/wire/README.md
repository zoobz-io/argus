# wire

Request and response types for the admin API surface.

## Purpose

Define the shape of data at the admin API boundary. Admin wire types expose full data without masking — internal team members need visibility into all fields for support and operations.

## Response Pattern

```go
// admin/wire/users.go
package wire

import "time"

// AdminUserResponse is the admin API response for user data.
// Unlike the public API, fields are not masked.
type AdminUserResponse struct {
    ID          int64     `json:"id" description:"User ID" example:"12345"`
    Login       string    `json:"login" description:"Username" example:"octocat"`
    Email       string    `json:"email" description:"Email address" example:"user@example.com"`
    Name        *string   `json:"name,omitempty" description:"Display name" example:"Jane Doe"`
    AvatarURL   *string   `json:"avatar_url,omitempty" description:"Profile image URL"`
    CreatedAt   time.Time `json:"created_at" description:"Account creation timestamp"`
    UpdatedAt   time.Time `json:"updated_at" description:"Last update timestamp"`
    LastLoginAt time.Time `json:"last_login_at" description:"Last login timestamp"`
    Status      string    `json:"status" description:"Account status" example:"active"`
}

// Clone returns a deep copy.
func (u AdminUserResponse) Clone() AdminUserResponse {
    c := u
    if u.Name != nil {
        n := *u.Name
        c.Name = &n
    }
    return c
}
```

## List Response Pattern

List responses use cursor-based pagination:

```go
// AdminTenantListResponse is the admin API response for listing tenants.
type AdminTenantListResponse struct {
    Tenants []AdminTenantResponse `json:"tenants" description:"List of tenants"`
    Cursor  *int64                `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
    Limit   int                   `json:"limit" description:"Page size" example:"20"`
    HasMore bool                  `json:"has_more" description:"Whether more results exist"`
}

// Clone returns a deep copy.
func (r AdminTenantListResponse) Clone() AdminTenantListResponse {
    c := r
    if r.Tenants != nil {
        c.Tenants = make([]AdminTenantResponse, len(r.Tenants))
        copy(c.Tenants, r.Tenants)
    }
    if r.Cursor != nil {
        v := *r.Cursor
        c.Cursor = &v
    }
    return c
}
```

## Key Differences from Public API

| Aspect | Public (api/wire) | Admin (admin/wire) |
|--------|-------------------|-------------------|
| Email | Masked (`send.mask:"email"`) | Exposed |
| Name | Masked (`send.mask:"name"`) | Exposed |
| Audit fields | Not included | Included (created_at, updated_at) |
| Internal data | Not included | Included (last_login, status) |
| OnSend boundary | Required for masking | Not needed |

## Guidelines

- No `send.mask` or `send.redact` tags — admin sees everything
- Include audit fields (created_at, updated_at, etc.)
- Include internal data (last_login, status, flags)
- Include pagination metadata for list responses
- Still implement `Clone()` if the type has pointer fields
- No `RegisterBoundaries()` needed — no masking required
