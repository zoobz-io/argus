package wire

import (
	"context"

	"github.com/zoobz-io/argus/internal/validate"
	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// TagResponse is the public API response for a tag.
type TagResponse struct {
	Name        string `json:"name" description:"Tag name" example:"compliance"`
	Description string `json:"description" description:"Tag description" example:"Compliance-related content"`
	ID          string `json:"id" description:"Tag ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OnSend applies boundary masking before the response is marshaled.
func (t *TagResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[TagResponse]](ctx)
	masked, err := b.Send(ctx, *t)
	if err != nil {
		return err
	}
	*t = masked
	return nil
}

// Clone returns a copy of the response.
func (t TagResponse) Clone() TagResponse {
	return t
}

// TagCreateRequest is the request body for creating or updating a tag.
type TagCreateRequest struct {
	Name        string `json:"name" description:"Tag name" example:"compliance"`
	Description string `json:"description" description:"Tag description" example:"Compliance-related content"`
}

// Validate validates the request fields using prompt-injection-safe rules.
func (r *TagCreateRequest) Validate() error {
	return check.All(
		validate.Name(r.Name, "name"),
		validate.Description(r.Description, "description"),
	).Err()
}

// Clone returns a copy of the request.
func (r TagCreateRequest) Clone() TagCreateRequest {
	return r
}
