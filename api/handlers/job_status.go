package handlers

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
)

// JobStatusSSE is the SSE event payload sent to clients.
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

var jobStatusStream = rocco.NewStreamHandler[rocco.NoBody, JobStatusSSE](
	"job-status",
	"GET",
	"/jobs/{id}/status",
	func(r *rocco.Request[rocco.NoBody], stream rocco.Stream[JobStatusSSE]) error {
		jobID := pathID(r.Params, "id")

		// 1. Read current state from DB.
		reader := sum.MustUse[contracts.JobReader](r)
		job, err := reader.GetJob(r, jobID)
		if err != nil {
			return ErrJobNotFound
		}

		// 2. Send current status as initial event.
		if err := stream.SendEvent("status", JobStatusSSE{
			JobID:  job.ID,
			Status: string(job.Status),
		}); err != nil {
			return err
		}

		// 3. If already terminal, send done and close.
		if job.Status == models.JobCompleted || job.Status == models.JobFailed {
			errMsg := ""
			if job.Error != nil {
				errMsg = *job.Error
			}
			return stream.SendEvent("done", JobStatusSSE{
				JobID:  job.ID,
				Status: string(job.Status),
				Error:  errMsg,
			})
		}

		// 4. Subscribe to live updates via capitan hook.
		listener := capitan.Hook(events.JobStatusSignal, func(_ context.Context, e *capitan.Event) {
			evt, ok := events.JobStatusKey.From(e)
			if !ok || evt.JobID != jobID {
				return
			}

			_ = stream.SendEvent(evt.Stage, JobStatusSSE{
				JobID:  evt.JobID,
				Status: evt.Stage,
				Stage:  evt.Stage,
				Error:  evt.Error,
			})

			// Close on terminal.
			if evt.Stage == "completed" || evt.Stage == "failed" {
				_ = stream.SendEvent("done", JobStatusSSE{
					JobID:  evt.JobID,
					Status: evt.Stage,
					Error:  evt.Error,
				})
			}
		})
		if listener != nil {
			defer listener.Close()
		}

		// 5. Block until client disconnects.
		<-stream.Done()
		return nil
	},
).WithPathParams("id").WithAuthentication()
