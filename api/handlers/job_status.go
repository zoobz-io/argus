package handlers

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/rocco"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/api/contracts"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
)

const sanitizedErrorMessage = "ingestion failed"

var jobStatusStream = rocco.NewStreamHandler[rocco.NoBody, wire.JobStatusSSE](
	"job-status",
	"GET",
	"/jobs/{id}/status",
	func(r *rocco.Request[rocco.NoBody], stream rocco.Stream[wire.JobStatusSSE]) error {
		jobID := pathID(r.Params, "id")
		tid := tenantID(r.Identity)

		// 1. Subscribe to live updates BEFORE reading DB state.
		// This prevents the race where a job completes between DB read and subscribe.
		done := make(chan struct{})
		listener := capitan.Hook(events.JobStatusSignal, func(_ context.Context, e *capitan.Event) {
			evt, ok := events.JobStatusKey.From(e)
			if !ok || evt.JobID != jobID {
				return
			}

			err := stream.SendEvent(evt.Stage, wire.JobStatusSSE{
				JobID:  evt.JobID,
				Status: evt.Stage,
				Stage:  evt.Stage,
				Error:  sanitizeError(evt.Error),
			})
			if err != nil {
				return // Client disconnected.
			}

			if evt.Stage == "completed" || evt.Stage == "failed" {
				_ = stream.SendEvent("done", wire.JobStatusSSE{
					JobID:  evt.JobID,
					Status: evt.Stage,
					Error:  sanitizeError(evt.Error),
				})
				close(done)
			}
		})
		if listener != nil {
			defer listener.Close()
		}

		// 2. Read current state from DB (tenant-scoped).
		reader := sum.MustUse[contracts.JobReader](r)
		job, err := reader.GetJobByTenant(r, jobID, tid)
		if err != nil {
			return stream.SendEvent("error", wire.JobStatusSSE{
				JobID:  jobID,
				Status: "not_found",
				Error:  "job not found",
			})
		}

		// 3. Send current status as initial event.
		if err := stream.SendEvent("status", wire.JobStatusSSE{
			JobID:  job.ID,
			Status: string(job.Status),
		}); err != nil {
			return err
		}

		// 4. If already terminal, send done and close.
		if job.Status == models.JobCompleted || job.Status == models.JobFailed {
			return stream.SendEvent("done", wire.JobStatusSSE{
				JobID:  job.ID,
				Status: string(job.Status),
				Error:  sanitizeJobError(job),
			})
		}

		// 5. Block until terminal event or client disconnect.
		select {
		case <-done:
		case <-stream.Done():
		}
		return nil
	},
).WithPathParams("id").WithAuthentication()

// sanitizeError returns a generic message for non-empty errors.
func sanitizeError(raw string) string {
	if raw == "" {
		return ""
	}
	return sanitizedErrorMessage
}

// sanitizeJobError returns a generic message if the job has an error.
func sanitizeJobError(job *models.Job) string {
	if job.Error == nil || *job.Error == "" {
		return ""
	}
	return sanitizedErrorMessage
}
