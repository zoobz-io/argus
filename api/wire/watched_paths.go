package wire

import (
	"context"

	"github.com/zoobz-io/check"
	"github.com/zoobz-io/sum"
)

// WatchedPathResponse is the API wire type for a watched path.
type WatchedPathResponse struct {
	Path       string `json:"path" description:"Watched path" example:"/documents/reports"`
	ID         int64  `json:"id" description:"Watched path ID" example:"1"`
	ProviderID int64  `json:"provider_id" description:"Associated provider ID" example:"1"`
	Active     bool   `json:"active" description:"Whether the path is being watched"`
}

// OnSend applies boundary rules before sending the response.
func (w *WatchedPathResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[WatchedPathResponse]](ctx)
	masked, err := b.Send(ctx, *w)
	if err != nil {
		return err
	}
	*w = masked
	return nil
}

// Clone returns a deep copy of the response.
func (w WatchedPathResponse) Clone() WatchedPathResponse {
	return w
}

// WatchedPathCreateRequest is the API wire type for creating a watched path.
type WatchedPathCreateRequest struct {
	Path       string `json:"path" description:"Path to watch" example:"/documents/reports"`
	ProviderID int64  `json:"provider_id" description:"Provider ID" example:"1"`
}

// Validate checks that all required fields are present and valid.
func (r *WatchedPathCreateRequest) Validate() error {
	return check.All(
		check.Str(r.Path, "path").Required().V(),
	).Err()
}

// Clone returns a deep copy of the request.
func (r WatchedPathCreateRequest) Clone() WatchedPathCreateRequest {
	return r
}
