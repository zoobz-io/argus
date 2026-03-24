package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// DocumentVersionListResponse is the API response for a paginated list of document versions.
type DocumentVersionListResponse struct {
	Cursor   *int64                    `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Versions []DocumentVersionResponse `json:"versions" description:"List of document versions"`
	Limit    int                       `json:"limit" description:"Page size" example:"20"`
	HasMore  bool                      `json:"has_more" description:"Whether more results exist"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *DocumentVersionListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[DocumentVersionListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r DocumentVersionListResponse) Clone() DocumentVersionListResponse {
	c := r
	if r.Versions != nil {
		c.Versions = make([]DocumentVersionResponse, len(r.Versions))
		copy(c.Versions, r.Versions)
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
