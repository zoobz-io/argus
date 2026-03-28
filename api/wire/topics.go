package wire

import (
	"context"

	"github.com/zoobz-io/argus/internal/validate"
	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// TopicResponse is the public API response for a topic.
type TopicResponse struct {
	Name        string `json:"name" description:"Topic name" example:"Security"`
	Description string `json:"description" description:"Topic description" example:"Security-related documents"`
	ID          string `json:"id" description:"Topic ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OnSend applies boundary masking before the response is marshaled.
func (t *TopicResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[TopicResponse]](ctx)
	masked, err := b.Send(ctx, *t)
	if err != nil {
		return err
	}
	*t = masked
	return nil
}

// Clone returns a copy of the response.
func (t TopicResponse) Clone() TopicResponse {
	return t
}

// TopicCreateRequest is the request body for creating or updating a topic.
type TopicCreateRequest struct {
	Name        string `json:"name" description:"Topic name" example:"Security"`
	Description string `json:"description" description:"Topic description" example:"Security-related documents"`
}

// Validate validates the request fields using prompt-injection-safe rules.
func (r *TopicCreateRequest) Validate() error {
	return check.All(
		validate.Name(r.Name, "name"),
		validate.Description(r.Description, "description"),
	).Err()
}

// Clone returns a copy of the request.
func (r TopicCreateRequest) Clone() TopicCreateRequest {
	return r
}
