package events

import "testing"

func TestFetchMessage_Clone(t *testing.T) {
	orig := FetchMessage{
		VersionID:  "v1",
		DocumentID: "d1",
		ProviderID: "p1",
		TenantID:   "t1",
		Ref:        "ref-abc",
		ObjectKey:  "key-xyz",
	}
	if orig.Clone() != orig {
		t.Error("clone mismatch")
	}
}
