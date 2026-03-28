package wire

import "github.com/zoobz-io/check"

// OffsetParams holds offset-based pagination parameters.
type OffsetParams struct {
	Offset int `json:"offset" description:"Number of results to skip" example:"0"`
	Limit  int `json:"limit" description:"Maximum results to return" example:"20"`
}

// Validate checks that pagination parameters are within acceptable bounds.
func (p *OffsetParams) Validate() error {
	return check.All(
		check.Int(p.Offset, "offset").Min(0).V(),
		check.Int(p.Limit, "limit").Min(1).Max(100).V(),
	).Err()
}

// Clone returns a copy of the params.
func (p OffsetParams) Clone() OffsetParams {
	return p
}
