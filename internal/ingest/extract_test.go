//go:build testing

package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ingest/extract"
	"github.com/zoobz-io/argus/models"
)

// --- tests for extractRoute ---

func TestExtractRoute_HappyPath(t *testing.T) {
	id := pipz.NewIdentity("test-extract", "test extractor")
	fn := func(_ context.Context, data []byte) (string, error) {
		return "extracted: " + string(data), nil
	}

	stage := extractRoute(id, fn)
	dc := &DocumentContext{
		RawBytes: []byte("hello"),
	}

	result, err := stage.Process(context.Background(), dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "extracted: hello" {
		t.Errorf("Content: got %q, want %q", result.Content, "extracted: hello")
	}
}

func TestExtractRoute_ExtractorError(t *testing.T) {
	id := pipz.NewIdentity("test-extract-fail", "failing extractor")
	fn := func(_ context.Context, _ []byte) (string, error) {
		return "", errors.New("parse error")
	}

	stage := extractRoute(id, fn)
	dc := &DocumentContext{
		RawBytes: []byte("bad data"),
	}

	_, err := stage.Process(context.Background(), dc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "parse error") {
		t.Errorf("error should contain 'parse error', got %q", err.Error())
	}
}

// --- tests for newExtractStage ---

func setupExtractRegistry(t *testing.T, versions intcontracts.IngestVersions) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	sum.Register[intcontracts.IngestVersions](k, versions)
	sum.Freeze(k)
	return context.Background()
}

func TestExtractStage_FetchAndPlaintext(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return nil, nil
		},
		OnGetVersionContent: func(_ context.Context, key string) ([]byte, error) {
			if key != "objects/doc-1/ver-1" {
				t.Errorf("object key: got %q, want %q", key, "objects/doc-1/ver-1")
			}
			return []byte("plain text content"), nil
		},
	}

	ctx := setupExtractRegistry(t, versions)
	stage := newExtractStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{
			ID:         "ver-1",
			DocumentID: "doc-1",
		},
		Document: &models.Document{
			ObjectKey: "objects/doc-1/ver-1",
			MimeType:  extract.MimePlain,
		},
		Job: &models.Job{ID: "job-1"},
	}

	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(result.RawBytes) != "plain text content" {
		t.Errorf("RawBytes: got %q, want %q", string(result.RawBytes), "plain text content")
	}
	if result.Content != "plain text content" {
		t.Errorf("Content: got %q, want %q", result.Content, "plain text content")
	}
}

func TestExtractStage_FetchError(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetDocumentVersion: func(_ context.Context, _ string) (*models.DocumentVersion, error) {
			return nil, nil
		},
		OnGetVersionContent: func(_ context.Context, _ string) ([]byte, error) {
			return nil, errors.New("storage unavailable")
		},
	}

	ctx := setupExtractRegistry(t, versions)
	stage := newExtractStage()

	dc := &DocumentContext{
		Version:  &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1"},
		Document: &models.Document{ObjectKey: "objects/doc-1/ver-1", MimeType: extract.MimePlain},
		Job:      &models.Job{ID: "job-1"},
	}

	_, err := stage.Process(ctx, dc)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "fetching object") {
		t.Errorf("error should contain 'fetching object', got %q", err.Error())
	}
}

func TestExtractStage_MarkdownRouting(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetVersionContent: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("# Heading\n\nBody text"), nil
		},
	}

	ctx := setupExtractRegistry(t, versions)
	stage := newExtractStage()

	dc := &DocumentContext{
		Version:  &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1"},
		Document: &models.Document{ObjectKey: "key", MimeType: extract.MimeMarkdown},
		Job:      &models.Job{ID: "job-1"},
	}

	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "# Heading\n\nBody text" {
		t.Errorf("Content: got %q, want %q", result.Content, "# Heading\n\nBody text")
	}
}

func TestExtractStage_CSVRouting(t *testing.T) {
	versions := &mockIngestVersions{
		OnGetVersionContent: func(_ context.Context, _ string) ([]byte, error) {
			return []byte("a,b,c\n1,2,3"), nil
		},
	}

	ctx := setupExtractRegistry(t, versions)
	stage := newExtractStage()

	dc := &DocumentContext{
		Version:  &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1"},
		Document: &models.Document{ObjectKey: "key", MimeType: extract.MimeCSV},
		Job:      &models.Job{ID: "job-1"},
	}

	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "a,b,c\n1,2,3" {
		t.Errorf("Content: got %q, want %q", result.Content, "a,b,c\n1,2,3")
	}
}
