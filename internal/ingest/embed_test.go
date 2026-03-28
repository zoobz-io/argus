//go:build testing

package ingest

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/sum"
	"github.com/zoobz-io/vex"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

// --- mocks ---

type mockEmbedder struct {
	OnEmbed func(ctx context.Context, text string) (vex.Vector, error)
}

func (m *mockEmbedder) Embed(ctx context.Context, text string) (vex.Vector, error) {
	return m.OnEmbed(ctx, text)
}

// --- helpers ---

func setupEmbedRegistry(t *testing.T, embedder intcontracts.Embedder) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	sum.Register[intcontracts.Embedder](k, embedder)
	sum.Freeze(k)
	return context.Background()
}

// --- tests ---

func TestEmbedStage_HappyPath(t *testing.T) {
	expected := vex.Vector{0.1, 0.2, 0.3}
	embedder := &mockEmbedder{
		OnEmbed: func(_ context.Context, text string) (vex.Vector, error) {
			if text != "test content" {
				t.Errorf("text: got %q, want %q", text, "test content")
			}
			return expected, nil
		},
	}

	ctx := setupEmbedRegistry(t, embedder)
	stage := newEmbedStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1"},
		Content: "test content",
	}

	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Embedding) != len(expected) {
		t.Fatalf("Embedding len: got %d, want %d", len(result.Embedding), len(expected))
	}
	for i := range expected {
		if result.Embedding[i] != expected[i] {
			t.Errorf("Embedding[%d]: got %f, want %f", i, result.Embedding[i], expected[i])
		}
	}
}

func TestEmbedStage_EmbedderError(t *testing.T) {
	embedder := &mockEmbedder{
		OnEmbed: func(_ context.Context, _ string) (vex.Vector, error) {
			return nil, errors.New("embedding failed")
		},
	}

	ctx := setupEmbedRegistry(t, embedder)
	stage := newEmbedStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1"},
		Content: "test content",
	}

	// Enrich swallows errors and returns dc unchanged.
	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("Enrich should not propagate errors, got: %v", err)
	}
	if result.Embedding != nil {
		t.Errorf("Embedding should be nil on failure, got %v", result.Embedding)
	}
}
