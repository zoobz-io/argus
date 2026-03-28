package wire

import "time"

// AdminTagResponse is the wire representation of a single tag.
type AdminTagResponse struct {
	CreatedAt   time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt   time.Time `json:"updated_at" description:"Last update timestamp"`
	Name        string    `json:"name" description:"Tag name" example:"compliance"`
	Description string    `json:"description" description:"Tag description" example:"Compliance-related content"`
	ID          string    `json:"id" description:"Tag ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	TenantID    string    `json:"tenant_id" description:"Owning tenant ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Clone returns a shallow copy of the response.
func (t AdminTagResponse) Clone() AdminTagResponse {
	return t
}

// AdminTagListResponse is the wire representation of a tag list.
type AdminTagListResponse struct {
	Tags []AdminTagResponse `json:"tags" description:"List of tags"`
}

// Clone returns a deep copy of the response.
func (r AdminTagListResponse) Clone() AdminTagListResponse {
	c := r
	if r.Tags != nil {
		c.Tags = make([]AdminTagResponse, len(r.Tags))
		copy(c.Tags, r.Tags)
	}
	return c
}
