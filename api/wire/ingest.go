package wire

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

// JobStatusSSE is the SSE event payload for job status updates.
type JobStatusSSE struct {
	JobID  string `json:"job_id"`
	Status string `json:"status"`
	Stage  string `json:"stage,omitempty"`
	Error  string `json:"error,omitempty"`
}

// Clone returns a copy of the event.
func (e JobStatusSSE) Clone() JobStatusSSE {
	return e
}

// NotificationSSE is the SSE event payload for real-time notification hints.
type NotificationSSE struct {
	NotificationID string `json:"notification_id"`
	Type           string `json:"type"`
	Message        string `json:"message"`
}

// Clone returns a copy of the event.
func (e NotificationSSE) Clone() NotificationSSE {
	return e
}
