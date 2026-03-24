package wire

import "github.com/zoobz-io/check"

// CursorParams holds cursor-based pagination parameters for admin endpoints.
type CursorParams struct {
	Cursor *int64 `json:"cursor,omitempty" description:"ID to start after (exclusive)" example:"42"`
	Limit  int    `json:"limit" description:"Maximum results to return" example:"20"`
}

// Validate checks that pagination parameters are within acceptable bounds.
func (p *CursorParams) Validate() error {
	return check.All(
		check.Int(p.Limit, "limit").Min(1).Max(100).V(),
	).Err()
}

// Clone returns a copy of the params.
func (p CursorParams) Clone() CursorParams {
	c := p
	if p.Cursor != nil {
		v := *p.Cursor
		c.Cursor = &v
	}
	return c
}
