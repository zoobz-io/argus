package wire

import "testing"

func TestIngestRequest_Clone(t *testing.T) {
	orig := IngestRequest{VersionID: "v-1"}
	cloned := orig.Clone()
	if cloned.VersionID != orig.VersionID {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestIngestResponse_Clone(t *testing.T) {
	orig := IngestResponse{JobID: "j-1", Status: "pending"}
	cloned := orig.Clone()
	if cloned.JobID != orig.JobID || cloned.Status != orig.Status {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestJobStatusSSE_Clone(t *testing.T) {
	orig := JobStatusSSE{JobID: "j-1", Status: "processing", Stage: "extract", Error: ""}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestNotificationSSE_Clone(t *testing.T) {
	orig := NotificationSSE{NotificationID: "n-1", Type: "ingest.completed", Message: "done"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}
