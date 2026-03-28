package models

import "testing"

func TestDocumentVersionIndex_Clone(t *testing.T) {
	d := DocumentVersionIndex{
		VersionID:    "ver-1",
		DocumentID:   "doc-1",
		TenantID:     "tenant-1",
		DocumentName: "report.pdf",
		Topics:       []string{"Finance", "Legal"},
		Tags:         []string{"Important"},
		Embedding:    []float32{0.1, 0.2, 0.3},
	}

	clone := d.Clone()

	// Fields copied.
	if clone.VersionID != d.VersionID {
		t.Errorf("VersionID: got %q, want %q", clone.VersionID, d.VersionID)
	}

	// Slice independence — Topics.
	clone.Topics[0] = "mutated"
	if d.Topics[0] != "Finance" {
		t.Error("Topics mutation leaked to original")
	}

	// Slice independence — Tags.
	clone.Tags[0] = "mutated"
	if d.Tags[0] != "Important" {
		t.Error("Tags mutation leaked to original")
	}

	// Slice independence — Embedding.
	clone.Embedding[0] = 999.0
	if d.Embedding[0] != 0.1 {
		t.Error("Embedding mutation leaked to original")
	}
}

func TestDocumentVersionIndex_Clone_NilSlices(t *testing.T) {
	d := DocumentVersionIndex{VersionID: "ver-1"}
	clone := d.Clone()

	if clone.Topics != nil {
		t.Error("expected nil Topics")
	}
	if clone.Tags != nil {
		t.Error("expected nil Tags")
	}
	if clone.Embedding != nil {
		t.Error("expected nil Embedding")
	}
}
