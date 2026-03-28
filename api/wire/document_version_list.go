package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// DocumentVersionListResponse is the API response for a paginated list of document versions.
type DocumentVersionListResponse struct {
	Versions []DocumentVersionResponse `json:"versions" description:"List of document versions"`
	Offset   int                       `json:"offset" description:"Number of results skipped"`
	Limit    int                       `json:"limit" description:"Page size" example:"20"`
	Total    int64                     `json:"total" description:"Total number of results"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *DocumentVersionListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[DocumentVersionListResponse]](ctx)
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
	return c
}
