//go:build testing

package handlers

import "testing"

func TestJobStatusSSE_Clone(t *testing.T) {
	orig := JobStatusSSE{JobID: "j1", Status: "pending", Stage: "started", Error: ""}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}
