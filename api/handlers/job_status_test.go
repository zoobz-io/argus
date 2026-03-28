//go:build testing

package handlers

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/zoobz-io/capitan"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/events"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// --- Clone ---

func TestJobStatusSSE_Clone(t *testing.T) {
	orig := JobStatusSSE{JobID: "j1", Status: "pending", Stage: "started", Error: ""}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

// --- Sanitize helpers ---

func TestSanitizeError_Empty(t *testing.T) {
	if got := sanitizeError(""); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSanitizeError_NonEmpty(t *testing.T) {
	got := sanitizeError("indexing version: connection refused to opensearch:9200")
	if got != sanitizedErrorMessage {
		t.Errorf("expected sanitized message, got %q", got)
	}
}

func TestSanitizeJobError_NilError(t *testing.T) {
	job := &models.Job{Error: nil}
	if got := sanitizeJobError(job); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSanitizeJobError_EmptyError(t *testing.T) {
	empty := ""
	job := &models.Job{Error: &empty}
	if got := sanitizeJobError(job); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestSanitizeJobError_NonEmpty(t *testing.T) {
	msg := "pipeline failed: embed: connection to api.openai.com timed out"
	job := &models.Job{Error: &msg}
	got := sanitizeJobError(job)
	if got != sanitizedErrorMessage {
		t.Errorf("expected sanitized message, got %q", got)
	}
}

// --- SSE handler tests ---

func TestJobStatus_NotFound(t *testing.T) {
	mock := &argustest.MockJobReader{
		OnGetJobByTenant: func(_ context.Context, _, _ string) (*models.Job, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIJobReader(mock),
		argustest.WithAPIIngestEnqueuer(&argustest.MockIngestEnqueuer{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeStream(engine, "GET", "/jobs/bad-id/status", nil)
	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	if len(evts) != 1 {
		t.Fatalf("expected 1 error event, got %d", len(evts))
	}
	if evts[0].Event != "error" {
		t.Errorf("expected event type %q, got %q", "error", evts[0].Event)
	}
	var evt JobStatusSSE
	if err := evts[0].DecodeJSON(&evt); err != nil {
		t.Fatalf("failed to decode error event: %v", err)
	}
	if evt.Status != "not_found" {
		t.Errorf("expected status %q, got %q", "not_found", evt.Status)
	}
}

func TestJobStatus_AlreadyCompleted(t *testing.T) {
	mock := &argustest.MockJobReader{
		OnGetJobByTenant: func(_ context.Context, id, _ string) (*models.Job, error) {
			return &models.Job{ID: id, Status: models.JobCompleted, TenantID: "tenant-1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIJobReader(mock),
		argustest.WithAPIIngestEnqueuer(&argustest.MockIngestEnqueuer{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeStream(engine, "GET", "/jobs/job-1/status", nil)

	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	if len(evts) != 2 {
		t.Fatalf("expected 2 events (status + done), got %d", len(evts))
	}
	if evts[0].Event != "status" {
		t.Errorf("event 0: expected type %q, got %q", "status", evts[0].Event)
	}
	if evts[1].Event != "done" {
		t.Errorf("event 1: expected type %q, got %q", "done", evts[1].Event)
	}

	var status JobStatusSSE
	if err := evts[0].DecodeJSON(&status); err != nil {
		t.Fatalf("failed to decode status event: %v", err)
	}
	if status.JobID != "job-1" {
		t.Errorf("status job_id: got %q, want %q", status.JobID, "job-1")
	}
	if status.Status != "completed" {
		t.Errorf("status: got %q, want %q", status.Status, "completed")
	}
}

func TestJobStatus_AlreadyFailed_ErrorSanitized(t *testing.T) {
	errMsg := "pipeline failed: connection refused to opensearch:9200"
	mock := &argustest.MockJobReader{
		OnGetJobByTenant: func(_ context.Context, id, _ string) (*models.Job, error) {
			return &models.Job{ID: id, Status: models.JobFailed, TenantID: "tenant-1", Error: &errMsg}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIJobReader(mock),
		argustest.WithAPIIngestEnqueuer(&argustest.MockIngestEnqueuer{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeStream(engine, "GET", "/jobs/job-1/status", nil)
	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	if len(evts) != 2 {
		t.Fatalf("expected 2 events, got %d", len(evts))
	}

	var done JobStatusSSE
	if err := evts[1].DecodeJSON(&done); err != nil {
		t.Fatalf("failed to decode done event: %v", err)
	}
	if done.Error != sanitizedErrorMessage {
		t.Errorf("error should be sanitized, got %q", done.Error)
	}
}

func TestJobStatus_LiveUpdates(t *testing.T) {
	mock := &argustest.MockJobReader{
		OnGetJobByTenant: func(_ context.Context, id, _ string) (*models.Job, error) {
			return &models.Job{ID: id, Status: models.JobProcessing, TenantID: "tenant-1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIJobReader(mock),
		argustest.WithAPIIngestEnqueuer(&argustest.MockIngestEnqueuer{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var capture *rtesting.StreamCapture
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		capture = rtesting.ServeStreamWithContext(ctx, engine, "GET", "/jobs/job-1/status", nil)
	}()

	// Give the handler time to subscribe and read DB.
	time.Sleep(50 * time.Millisecond)

	// Emit a stage signal — the handler's capitan hook should pick it up.
	capitan.Emit(context.Background(), events.JobStatusSignal, events.JobStatusKey.Field(events.JobStatusEvent{
		JobID: "job-1",
		Stage: "extracted",
	}))

	// Emit terminal signal.
	capitan.Emit(context.Background(), events.JobStatusSignal, events.JobStatusKey.Field(events.JobStatusEvent{
		JobID: "job-1",
		Stage: "completed",
	}))

	// Wait for handler to finish (done channel closed by terminal event).
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		cancel() // Unblock if stuck.
		wg.Wait()
		t.Fatal("handler did not complete after terminal signal")
	}

	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	// Expect: status (initial), extracted, completed, done
	if len(evts) < 3 {
		t.Fatalf("expected at least 3 events, got %d", len(evts))
	}

	// First event is the initial status.
	if evts[0].Event != "status" {
		t.Errorf("event 0: expected %q, got %q", "status", evts[0].Event)
	}
}

func TestJobStatus_IgnoresOtherJobs(t *testing.T) {
	mock := &argustest.MockJobReader{
		OnGetJobByTenant: func(_ context.Context, id, _ string) (*models.Job, error) {
			return &models.Job{ID: id, Status: models.JobProcessing, TenantID: "tenant-1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIJobReader(mock),
		argustest.WithAPIIngestEnqueuer(&argustest.MockIngestEnqueuer{}),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var capture *rtesting.StreamCapture
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		capture = rtesting.ServeStreamWithContext(ctx, engine, "GET", "/jobs/job-1/status", nil)
	}()

	time.Sleep(50 * time.Millisecond)

	// Emit signal for a DIFFERENT job — should be filtered out.
	capitan.Emit(context.Background(), events.JobStatusSignal, events.JobStatusKey.Field(events.JobStatusEvent{
		JobID: "job-other",
		Stage: "extracted",
	}))

	// Emit terminal for the right job.
	capitan.Emit(context.Background(), events.JobStatusSignal, events.JobStatusKey.Field(events.JobStatusEvent{
		JobID: "job-1",
		Stage: "completed",
	}))

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		cancel()
		wg.Wait()
		t.Fatal("handler did not complete")
	}

	rtesting.AssertSSE(t, capture)

	evts := capture.ParseEvents()
	// Should NOT contain an "extracted" event (that was for job-other).
	for _, evt := range evts {
		if evt.Event == "extracted" {
			t.Error("handler should not have forwarded event for different job")
		}
	}
}
