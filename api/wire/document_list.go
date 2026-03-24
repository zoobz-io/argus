package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// DocumentListResponse is the API response for a paginated list of documents.
type DocumentListResponse struct {
	Cursor    *int64             `json:"cursor,omitempty" description:"Cursor for next page (last ID in this page)"`
	Documents []DocumentResponse `json:"documents" description:"List of documents"`
	Limit     int                `json:"limit" description:"Page size" example:"20"`
	HasMore   bool               `json:"has_more" description:"Whether more results exist"`
}

// OnSend applies boundary masking before the response is marshaled.
func (r *DocumentListResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[DocumentListResponse]](ctx)
	masked, err := b.Send(ctx, *r)
	if err != nil {
		return err
	}
	*r = masked
	return nil
}

// Clone returns a deep copy of the response.
func (r DocumentListResponse) Clone() DocumentListResponse {
	c := r
	if r.Documents != nil {
		c.Documents = make([]DocumentResponse, len(r.Documents))
		for i, d := range r.Documents {
			c.Documents[i] = d.Clone()
		}
	}
	if r.Cursor != nil {
		v := *r.Cursor
		c.Cursor = &v
	}
	return c
}
