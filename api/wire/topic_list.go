package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// TopicListResponse is the API response for a list of topics.
type TopicListResponse struct {
	Topics []TopicResponse `json:"topics" description:"List of topics"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *TopicListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[TopicListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r TopicListResponse) Clone() TopicListResponse {
	c := r
	if r.Topics != nil {
		c.Topics = make([]TopicResponse, len(r.Topics))
		copy(c.Topics, r.Topics)
	}
	return c
}
