package handlers

import (
	"log"

	"github.com/zoobz-io/rocco"
)

// IngestRequest is the request body for triggering manual ingestion.
type IngestRequest struct {
	DocumentID int64 `json:"document_id" description:"Document to ingest" example:"1"`
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
	log.Printf("manual ingestion triggered for document_id=%d", r.Body.DocumentID)
	return IngestResponse{Message: "ingestion triggered"}, nil
}).
	WithSummary("Trigger manual ingestion").
	WithTags("ingestion").
	WithAuthentication()
