package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// WatchedPathListResponse is the API response for a paginated list of watched paths.
type WatchedPathListResponse struct {
	Cursor       *int64                `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	WatchedPaths []WatchedPathResponse `json:"watched_paths" description:"List of watched paths"`
	Limit        int                   `json:"limit" description:"Page size" example:"20"`
	HasMore      bool                  `json:"has_more" description:"Whether more results exist"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *WatchedPathListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[WatchedPathListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r WatchedPathListResponse) Clone() WatchedPathListResponse {
	c := r
	if r.WatchedPaths != nil {
		c.WatchedPaths = make([]WatchedPathResponse, len(r.WatchedPaths))
		copy(c.WatchedPaths, r.WatchedPaths)
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
