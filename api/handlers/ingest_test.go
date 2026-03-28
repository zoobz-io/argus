//go:build testing

package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestIngestRequest_Clone(t *testing.T) {
	orig := IngestRequest{VersionID: "v1"}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

func TestIngestResponse_Clone(t *testing.T) {
	orig := IngestResponse{Message: "done"}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

func TestTriggerIngest_Success(t *testing.T) {
	var capturedVersionID string
	mock := &argustest.MockIngest{
		OnIngest: func(_ context.Context, versionID string) error {
			capturedVersionID = versionID
			return nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIIngest(mock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := IngestRequest{VersionID: "v1"}
	capture := rtesting.ServeRequest(engine, "POST", "/ingest", body)
	rtesting.AssertStatus(t, capture, 200)

	if capturedVersionID != "v1" {
		t.Errorf("expected v1, got %q", capturedVersionID)
	}
}

func TestTriggerIngest_Error(t *testing.T) {
	mock := &argustest.MockIngest{
		OnIngest: func(_ context.Context, _ string) error {
			return fmt.Errorf("pipeline failed")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(), argustest.WithAPIIngest(mock), argustest.WithBoundaries(wire.RegisterBoundaries))

	body := IngestRequest{VersionID: "v1"}
	capture := rtesting.ServeRequest(engine, "POST", "/ingest", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error status")
	}
}
