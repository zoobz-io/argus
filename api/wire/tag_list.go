package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// TagListResponse is the API response for a list of tags.
type TagListResponse struct {
	Tags []TagResponse `json:"tags" description:"List of tags"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *TagListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[TagListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r TagListResponse) Clone() TagListResponse {
	c := r
	if r.Tags != nil {
		c.Tags = make([]TagResponse, len(r.Tags))
		copy(c.Tags, r.Tags)
	}
	return c
}
