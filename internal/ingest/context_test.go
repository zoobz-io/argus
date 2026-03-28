package ingest

import (
	"testing"
	"time"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/vex"
)

func TestDocumentContext_Clone_FullCopy(t *testing.T) {
	versionID := "ver-1"
	dc := &DocumentContext{
		Version: &models.DocumentVersion{
			ID:         versionID,
			DocumentID: "doc-1",
			TenantID:   "tenant-1",
		},
		Document: &models.Document{
			ID:       "doc-1",
			Name:     "test.pdf",
			MimeType: "application/pdf",
		},
		Job: &models.Job{
			ID:     "job-1",
			Status: models.JobProcessing,
		},
		Topics:    []string{"finance", "legal"},
		Tags:      []string{"important", "review"},
		RawBytes:  []byte("raw content"),
		Content:   "extracted text",
		Summary:   "a summary",
		Language:  "en",
		Embedding: vex.Vector{0.1, 0.2, 0.3},
	}

	clone := dc.Clone()

	// Verify all scalar fields copied.
	if clone.Content != dc.Content {
		t.Errorf("Content: got %q, want %q", clone.Content, dc.Content)
	}
	if clone.Summary != dc.Summary {
		t.Errorf("Summary: got %q, want %q", clone.Summary, dc.Summary)
	}
	if clone.Language != dc.Language {
		t.Errorf("Language: got %q, want %q", clone.Language, dc.Language)
	}

	// Verify pointer fields are independent copies.
	if clone.Version == dc.Version {
		t.Error("Version: clone shares pointer with original")
	}
	if clone.Version.ID != versionID {
		t.Errorf("Version.ID: got %q, want %q", clone.Version.ID, versionID)
	}
	if clone.Document == dc.Document {
		t.Error("Document: clone shares pointer with original")
	}
	if clone.Job == dc.Job {
		t.Error("Job: clone shares pointer with original")
	}

	// Verify slice fields have equal content.
	if len(clone.Topics) != len(dc.Topics) {
		t.Fatalf("Topics len: got %d, want %d", len(clone.Topics), len(dc.Topics))
	}
	for i := range dc.Topics {
		if clone.Topics[i] != dc.Topics[i] {
			t.Errorf("Topics[%d]: got %q, want %q", i, clone.Topics[i], dc.Topics[i])
		}
	}
	if len(clone.Embedding) != len(dc.Embedding) {
		t.Fatalf("Embedding len: got %d, want %d", len(clone.Embedding), len(dc.Embedding))
	}
}

func TestDocumentContext_Clone_SliceIndependence(t *testing.T) {
	dc := &DocumentContext{
		Topics:    []string{"original"},
		Tags:      []string{"original"},
		RawBytes:  []byte("original"),
		Embedding: vex.Vector{1.0},
	}

	clone := dc.Clone()

	// Mutate clone slices.
	clone.Topics[0] = "mutated"
	clone.Tags[0] = "mutated"
	clone.RawBytes[0] = 0xFF
	clone.Embedding[0] = 999.0

	// Original should be unchanged.
	if dc.Topics[0] != "original" {
		t.Error("Topics mutation leaked to original")
	}
	if dc.Tags[0] != "original" {
		t.Error("Tags mutation leaked to original")
	}
	if dc.RawBytes[0] == 0xFF {
		t.Error("RawBytes mutation leaked to original")
	}
	if dc.Embedding[0] != 1.0 {
		t.Error("Embedding mutation leaked to original")
	}
}

func TestDocumentContext_Clone_PointerIndependence(t *testing.T) {
	dc := &DocumentContext{
		Version: &models.DocumentVersion{
			ID:       "ver-1",
			TenantID: "tenant-1",
		},
		Document: &models.Document{
			ID:   "doc-1",
			Name: "original.pdf",
		},
		Job: &models.Job{
			ID:     "job-1",
			Status: models.JobProcessing,
		},
	}

	clone := dc.Clone()

	// Mutate clone pointers.
	clone.Version.ID = "ver-mutated"
	clone.Document.Name = "mutated.pdf"
	clone.Job.Status = models.JobFailed

	// Original should be unchanged.
	if dc.Version.ID != "ver-1" {
		t.Error("Version mutation leaked to original")
	}
	if dc.Document.Name != "original.pdf" {
		t.Error("Document mutation leaked to original")
	}
	if dc.Job.Status != models.JobProcessing {
		t.Error("Job mutation leaked to original")
	}
}

func TestDocumentContext_Clone_NilFields(t *testing.T) {
	dc := &DocumentContext{
		Content:  "some content",
		Language: "en",
	}

	clone := dc.Clone()

	if clone.Version != nil {
		t.Error("Version: expected nil")
	}
	if clone.Document != nil {
		t.Error("Document: expected nil")
	}
	if clone.Job != nil {
		t.Error("Job: expected nil")
	}
	if clone.Topics != nil {
		t.Error("Topics: expected nil")
	}
	if clone.Tags != nil {
		t.Error("Tags: expected nil")
	}
	if clone.RawBytes != nil {
		t.Error("RawBytes: expected nil")
	}
	if clone.Embedding != nil {
		t.Error("Embedding: expected nil")
	}
	if clone.Content != "some content" {
		t.Errorf("Content: got %q, want %q", clone.Content, "some content")
	}
}

func TestDocumentContext_Clone_DocumentWithPointerField(t *testing.T) {
	cvid := "ver-current"
	dc := &DocumentContext{
		Document: &models.Document{
			ID:               "doc-1",
			CurrentVersionID: &cvid,
			CreatedAt:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	clone := dc.Clone()

	if clone.Document.CurrentVersionID == dc.Document.CurrentVersionID {
		t.Error("Document.CurrentVersionID: clone shares pointer with original")
	}
	if *clone.Document.CurrentVersionID != cvid {
		t.Errorf("Document.CurrentVersionID: got %q, want %q", *clone.Document.CurrentVersionID, cvid)
	}
}

func TestDocumentContext_Clone_JobWithErrorPointer(t *testing.T) {
	errMsg := "something failed"
	dc := &DocumentContext{
		Job: &models.Job{
			ID:    "job-1",
			Error: &errMsg,
		},
	}

	clone := dc.Clone()

	if clone.Job.Error == dc.Job.Error {
		t.Error("Job.Error: clone shares pointer with original")
	}
	if *clone.Job.Error != errMsg {
		t.Errorf("Job.Error: got %q, want %q", *clone.Job.Error, errMsg)
	}
}
