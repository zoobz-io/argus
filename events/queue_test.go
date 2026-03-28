package events

import "testing"

func TestIngestMessage_Clone(t *testing.T) {
	orig := IngestMessage{JobID: "j1", VersionID: "v1"}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}
