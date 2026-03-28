package events

import "github.com/zoobz-io/capitan"

// Job status signals for SSE streaming.
var (
	JobStatusSignal = capitan.NewSignal("argus.job.status", "Job status update for SSE streaming")
	JobStatusKey    = capitan.NewKey[JobStatusEvent]("job_status", "events.JobStatusEvent")
)

// JobStatusEvent is the payload published to the job-status herald stream.
type JobStatusEvent struct {
	JobID      string `json:"job_id"`
	VersionID  string `json:"version_id"`
	DocumentID string `json:"document_id"`
	TenantID   string `json:"tenant_id"`
	Stage      string `json:"stage"`
	Error      string `json:"error,omitempty"`
}

// Clone returns a deep copy.
func (e JobStatusEvent) Clone() JobStatusEvent {
	return e
}
