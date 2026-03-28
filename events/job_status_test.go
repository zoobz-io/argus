package events

import "testing"

func TestJobStatusEvent_Clone(t *testing.T) {
	orig := JobStatusEvent{
		JobID:      "j1",
		VersionID:  "v1",
		DocumentID: "d1",
		TenantID:   "t1",
		Stage:      "extracted",
		Error:      "",
	}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}
