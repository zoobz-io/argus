//go:build testing

package handlers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestJobStatusSSE_Clone(t *testing.T) {
	orig := JobStatusSSE{JobID: "j1", Status: "pending", Stage: "started", Error: ""}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}

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
