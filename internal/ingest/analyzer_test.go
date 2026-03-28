package ingest

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/zoobz-io/zyn"

	"github.com/zoobz-io/argus/models"
)

// --- mock ---

type mockExtractionSynapse struct {
	OnFireWithInput func(ctx context.Context, session *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error)
}

func (m *mockExtractionSynapse) FireWithInput(ctx context.Context, session *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error) {
	return m.OnFireWithInput(ctx, session, input)
}

// --- tests ---

func TestSynapseAnalyzer_Analyze_BothVocabularies(t *testing.T) {
	var capturedInput zyn.ExtractionInput
	synapse := &mockExtractionSynapse{
		OnFireWithInput: func(_ context.Context, _ *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error) {
			capturedInput = input
			return models.DocumentAnalysis{
				Summary:  "a summary",
				Language: "en",
				Topics:   []string{"Finance"},
				Tags:     []string{"Urgent"},
			}, nil
		},
	}

	analyzer := &SynapseAnalyzer{synapse: synapse}
	result, err := analyzer.Analyze(context.Background(), "doc content", []string{"Finance", "Legal"}, []string{"Urgent", "Review"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify prompt construction.
	if capturedInput.Text != "doc content" {
		t.Errorf("Text: got %q, want %q", capturedInput.Text, "doc content")
	}
	if !strings.Contains(capturedInput.Context, "Allowed topics (select ONLY from this list): Finance, Legal") {
		t.Errorf("Context should contain topic list, got %q", capturedInput.Context)
	}
	if !strings.Contains(capturedInput.Context, "Allowed tags (select ONLY from this list): Urgent, Review") {
		t.Errorf("Context should contain tag list, got %q", capturedInput.Context)
	}

	// Verify result passthrough.
	if result.Summary != "a summary" {
		t.Errorf("Summary: got %q, want %q", result.Summary, "a summary")
	}
	if result.Language != "en" {
		t.Errorf("Language: got %q, want %q", result.Language, "en")
	}
}

func TestSynapseAnalyzer_Analyze_EmptyTopics(t *testing.T) {
	var capturedInput zyn.ExtractionInput
	synapse := &mockExtractionSynapse{
		OnFireWithInput: func(_ context.Context, _ *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error) {
			capturedInput = input
			return models.DocumentAnalysis{Summary: "s", Language: "en"}, nil
		},
	}

	analyzer := &SynapseAnalyzer{synapse: synapse}
	_, err := analyzer.Analyze(context.Background(), "content", nil, []string{"Tag1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedInput.Context, "No topics are defined") {
		t.Errorf("Context should contain empty-topics fallback, got %q", capturedInput.Context)
	}
	if !strings.Contains(capturedInput.Context, "Allowed tags") {
		t.Errorf("Context should contain tag list, got %q", capturedInput.Context)
	}
}

func TestSynapseAnalyzer_Analyze_EmptyTags(t *testing.T) {
	var capturedInput zyn.ExtractionInput
	synapse := &mockExtractionSynapse{
		OnFireWithInput: func(_ context.Context, _ *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error) {
			capturedInput = input
			return models.DocumentAnalysis{Summary: "s", Language: "en"}, nil
		},
	}

	analyzer := &SynapseAnalyzer{synapse: synapse}
	_, err := analyzer.Analyze(context.Background(), "content", []string{"Topic1"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedInput.Context, "Allowed topics") {
		t.Errorf("Context should contain topic list, got %q", capturedInput.Context)
	}
	if !strings.Contains(capturedInput.Context, "No tags are defined") {
		t.Errorf("Context should contain empty-tags fallback, got %q", capturedInput.Context)
	}
}

func TestSynapseAnalyzer_Analyze_BothEmpty(t *testing.T) {
	var capturedInput zyn.ExtractionInput
	synapse := &mockExtractionSynapse{
		OnFireWithInput: func(_ context.Context, _ *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error) {
			capturedInput = input
			return models.DocumentAnalysis{Summary: "s", Language: "en"}, nil
		},
	}

	analyzer := &SynapseAnalyzer{synapse: synapse}
	_, err := analyzer.Analyze(context.Background(), "content", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(capturedInput.Context, "No topics are defined") {
		t.Errorf("Context should contain empty-topics fallback, got %q", capturedInput.Context)
	}
	if !strings.Contains(capturedInput.Context, "No tags are defined") {
		t.Errorf("Context should contain empty-tags fallback, got %q", capturedInput.Context)
	}
}

func TestSynapseAnalyzer_Analyze_SynapseError(t *testing.T) {
	synapse := &mockExtractionSynapse{
		OnFireWithInput: func(_ context.Context, _ *zyn.Session, _ zyn.ExtractionInput) (models.DocumentAnalysis, error) {
			return models.DocumentAnalysis{}, errors.New("llm timeout")
		},
	}

	analyzer := &SynapseAnalyzer{synapse: synapse}
	_, err := analyzer.Analyze(context.Background(), "content", nil, nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "analyzing document") {
		t.Errorf("error should wrap with 'analyzing document', got %q", err.Error())
	}
	if !strings.Contains(err.Error(), "llm timeout") {
		t.Errorf("error should contain cause, got %q", err.Error())
	}
}

func TestSynapseAnalyzer_Analyze_ContextNewlineSeparation(t *testing.T) {
	var capturedInput zyn.ExtractionInput
	synapse := &mockExtractionSynapse{
		OnFireWithInput: func(_ context.Context, _ *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error) {
			capturedInput = input
			return models.DocumentAnalysis{Summary: "s", Language: "en"}, nil
		},
	}

	analyzer := &SynapseAnalyzer{synapse: synapse}
	_, _ = analyzer.Analyze(context.Background(), "content", []string{"A"}, []string{"B"})

	parts := strings.Split(capturedInput.Context, "\n")
	if len(parts) != 2 {
		t.Errorf("Context should have 2 newline-separated parts, got %d: %q", len(parts), capturedInput.Context)
	}
}
