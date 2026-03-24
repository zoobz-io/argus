package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// IngestRequest is the request body for triggering manual ingestion.
type IngestRequest struct {
	VersionID int64 `json:"version_id" description:"Document version to ingest" example:"1"`
}

// Clone returns a copy of the request.
func (r IngestRequest) Clone() IngestRequest {
	return r
}

// IngestResponse is the response for a manual ingestion trigger.
type IngestResponse struct {
	Message string `json:"message" description:"Status message"`
}

// Clone returns a copy of the response.
func (r IngestResponse) Clone() IngestResponse {
	return r
}

var triggerIngest = rocco.POST[IngestRequest, IngestResponse]("/ingest", func(r *rocco.Request[IngestRequest]) (IngestResponse, error) {
	pipeline := sum.MustUse[contracts.Ingest](r)
	if err := pipeline.Ingest(r, r.Body.VersionID); err != nil {
		return IngestResponse{}, err
	}
	return IngestResponse{Message: "ingestion completed"}, nil
}).
	WithSummary("Trigger manual ingestion").
	WithTags("ingestion").
	WithAuthentication()
