package wire

import "time"

// AdminWatchedPathResponse is the wire representation of a single watched path.
type AdminWatchedPathResponse struct {
	CreatedAt  time.Time `json:"created_at" description:"Creation timestamp"`
	UpdatedAt  time.Time `json:"updated_at" description:"Last update timestamp"`
	Path       string    `json:"path" description:"Watched path" example:"/documents/reports"`
	ID         int64     `json:"id" description:"Watched path ID" example:"1"`
	TenantID   int64     `json:"tenant_id" description:"Owning tenant ID" example:"1"`
	ProviderID int64     `json:"provider_id" description:"Associated provider ID" example:"1"`
	Active     bool      `json:"active" description:"Whether the path is being watched"`
}

// Clone returns a shallow copy of the response.
func (w AdminWatchedPathResponse) Clone() AdminWatchedPathResponse {
	return w
}

// AdminWatchedPathListResponse is the wire representation of a paginated watched path list.
type AdminWatchedPathListResponse struct {
	WatchedPaths []AdminWatchedPathResponse `json:"watched_paths" description:"List of watched paths"`
	Total        int                        `json:"total" description:"Total matching paths"`
	Limit        int                        `json:"limit" description:"Page size"`
	Offset       int                        `json:"offset" description:"Page offset"`
}

// Clone returns a deep copy of the response.
func (r AdminWatchedPathListResponse) Clone() AdminWatchedPathListResponse {
	c := r
	if r.WatchedPaths != nil {
		c.WatchedPaths = make([]AdminWatchedPathResponse, len(r.WatchedPaths))
		copy(c.WatchedPaths, r.WatchedPaths)
	}
	return c
}
