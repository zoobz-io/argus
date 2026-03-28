package handlers

import (
	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"
)

// IngestRequest is the request body for triggering ingestion.
type IngestRequest struct {
	VersionID string `json:"version_id" description:"Document version to ingest" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Clone returns a copy of the request.
func (r IngestRequest) Clone() IngestRequest {
	return r
}

// IngestResponse is the response for an ingestion trigger.
type IngestResponse struct {
	JobID  string `json:"job_id" description:"Ingestion job identifier"`
	Status string `json:"status" description:"Current job status"`
}

// Clone returns a copy of the response.
func (r IngestResponse) Clone() IngestResponse {
	return r
}

var triggerIngest = rocco.POST[IngestRequest, IngestResponse]("/ingest", func(r *rocco.Request[IngestRequest]) (IngestResponse, error) {
	enqueuer := sum.MustUse[contracts.IngestEnqueuer](r)
	job, err := enqueuer.Enqueue(r, r.Body.VersionID, tenantID(r.Identity))
	if err != nil {
		return IngestResponse{}, err
	}
	return IngestResponse{
		JobID:  job.ID,
		Status: string(job.Status),
	}, nil
}).
	WithSummary("Trigger async ingestion").
	WithTags("ingestion").
	WithAuthentication()
