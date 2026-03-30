//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestIngestRequest_Clone(t *testing.T) {
	orig := wire.IngestRequest{VersionID: "v1"}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

func TestIngestResponse_Clone(t *testing.T) {
	orig := wire.IngestResponse{JobID: "j1", Status: "pending"}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

func TestTriggerIngest_Success(t *testing.T) {
	var capturedVersionID string
	mock := &argustest.MockIngestEnqueuer{
		OnEnqueue: func(_ context.Context, versionID, _ string) (*models.Job, error) {
			capturedVersionID = versionID
			return &models.Job{ID: "job-1", Status: models.JobPending}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIIngestEnqueuer(mock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.IngestRequest{VersionID: "v1"}
	capture := rtesting.ServeRequest(engine, "POST", "/ingest", body)
	rtesting.AssertStatus(t, capture, 200)

	if capturedVersionID != "v1" {
		t.Errorf("expected v1, got %q", capturedVersionID)
	}

	var resp wire.IngestResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.JobID != "job-1" {
		t.Errorf("expected job-1, got %q", resp.JobID)
	}
	if resp.Status != "pending" {
		t.Errorf("expected pending, got %q", resp.Status)
	}
}

func TestTriggerIngest_Error(t *testing.T) {
	mock := &argustest.MockIngestEnqueuer{
		OnEnqueue: func(_ context.Context, _, _ string) (*models.Job, error) {
			return nil, fmt.Errorf("pipeline failed")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIIngestEnqueuer(mock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := wire.IngestRequest{VersionID: "v1"}
	capture := rtesting.ServeRequest(engine, "POST", "/ingest", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}
