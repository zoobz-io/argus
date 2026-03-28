//go:build testing

package ingest

import (
	"context"
	"errors"
	"testing"

	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

// --- mocks ---

type mockAnalyzer struct {
	OnAnalyze func(ctx context.Context, content string, topics, tags []string) (*models.DocumentAnalysis, error)
}

func (m *mockAnalyzer) Analyze(ctx context.Context, content string, topics, tags []string) (*models.DocumentAnalysis, error) {
	return m.OnAnalyze(ctx, content, topics, tags)
}

type mockIngestTopics struct {
	OnListTopicsByTenant func(ctx context.Context, tenantID string) ([]*models.Topic, error)
}

func (m *mockIngestTopics) ListTopicsByTenant(ctx context.Context, tenantID string) ([]*models.Topic, error) {
	return m.OnListTopicsByTenant(ctx, tenantID)
}

type mockIngestTags struct {
	OnListTagsByTenant func(ctx context.Context, tenantID string) ([]*models.Tag, error)
}

func (m *mockIngestTags) ListTagsByTenant(ctx context.Context, tenantID string) ([]*models.Tag, error) {
	return m.OnListTagsByTenant(ctx, tenantID)
}

// --- helpers ---

func setupAnalyzeRegistry(t *testing.T, analyzer intcontracts.Analyzer, topics intcontracts.IngestTopics, tags intcontracts.IngestTags) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	sum.Register[intcontracts.Analyzer](k, analyzer)
	sum.Register[intcontracts.IngestTopics](k, topics)
	sum.Register[intcontracts.IngestTags](k, tags)
	sum.Freeze(k)
	return context.Background()
}

// --- tests ---

func TestAnalyzeStage_HappyPath(t *testing.T) {
	analyzer := &mockAnalyzer{
		OnAnalyze: func(_ context.Context, content string, topics, tags []string) (*models.DocumentAnalysis, error) {
			if content != "test content" {
				t.Errorf("content: got %q, want %q", content, "test content")
			}
			if len(topics) != 2 || topics[0] != "Finance" || topics[1] != "Legal" {
				t.Errorf("topics: got %v, want [Finance Legal]", topics)
			}
			if len(tags) != 1 || tags[0] != "Important" {
				t.Errorf("tags: got %v, want [Important]", tags)
			}
			return &models.DocumentAnalysis{
				Summary:  "a summary",
				Language: "en",
				Topics:   []string{"Finance"},
				Tags:     []string{"Important"},
			}, nil
		},
	}
	topicStore := &mockIngestTopics{
		OnListTopicsByTenant: func(_ context.Context, tenantID string) ([]*models.Topic, error) {
			return []*models.Topic{
				{Name: "Finance", TenantID: tenantID},
				{Name: "Legal", TenantID: tenantID},
			}, nil
		},
	}
	tagStore := &mockIngestTags{
		OnListTagsByTenant: func(_ context.Context, tenantID string) ([]*models.Tag, error) {
			return []*models.Tag{
				{Name: "Important", TenantID: tenantID},
			}, nil
		},
	}

	ctx := setupAnalyzeRegistry(t, analyzer, topicStore, tagStore)
	stage := newAnalyzeStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"},
		Job:     &models.Job{ID: "job-1"},
		Content: "test content",
	}

	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "a summary" {
		t.Errorf("Summary: got %q, want %q", result.Summary, "a summary")
	}
	if result.Language != "en" {
		t.Errorf("Language: got %q, want %q", result.Language, "en")
	}
	if len(result.Topics) != 1 || result.Topics[0] != "Finance" {
		t.Errorf("Topics: got %v, want [Finance]", result.Topics)
	}
	if len(result.Tags) != 1 || result.Tags[0] != "Important" {
		t.Errorf("Tags: got %v, want [Important]", result.Tags)
	}
}

func TestAnalyzeStage_TopicStoreError(t *testing.T) {
	topicStore := &mockIngestTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) {
			return nil, errors.New("topic store failure")
		},
	}
	tagStore := &mockIngestTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) {
			return nil, nil
		},
	}
	analyzer := &mockAnalyzer{
		OnAnalyze: func(_ context.Context, _ string, _, _ []string) (*models.DocumentAnalysis, error) {
			t.Error("analyzer should not be called when topic store fails")
			return nil, nil
		},
	}

	ctx := setupAnalyzeRegistry(t, analyzer, topicStore, tagStore)
	stage := newAnalyzeStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"},
		Job:     &models.Job{ID: "job-1"},
		Content: "test content",
	}

	// Enrich swallows errors and returns dc unchanged.
	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("Enrich should not propagate errors, got: %v", err)
	}
	if result.Summary != "" {
		t.Errorf("Summary should be empty on failure, got %q", result.Summary)
	}
}

func TestAnalyzeStage_TagStoreError(t *testing.T) {
	topicStore := &mockIngestTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) {
			return nil, nil
		},
	}
	tagStore := &mockIngestTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) {
			return nil, errors.New("tag store failure")
		},
	}
	analyzer := &mockAnalyzer{
		OnAnalyze: func(_ context.Context, _ string, _, _ []string) (*models.DocumentAnalysis, error) {
			t.Error("analyzer should not be called when tag store fails")
			return nil, nil
		},
	}

	ctx := setupAnalyzeRegistry(t, analyzer, topicStore, tagStore)
	stage := newAnalyzeStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"},
		Job:     &models.Job{ID: "job-1"},
		Content: "test content",
	}

	// Enrich swallows errors and returns dc unchanged.
	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("Enrich should not propagate errors, got: %v", err)
	}
	if result.Summary != "" {
		t.Errorf("Summary should be empty on failure, got %q", result.Summary)
	}
}

func TestAnalyzeStage_AnalyzerError(t *testing.T) {
	topicStore := &mockIngestTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) {
			return nil, nil
		},
	}
	tagStore := &mockIngestTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) {
			return nil, nil
		},
	}
	analyzer := &mockAnalyzer{
		OnAnalyze: func(_ context.Context, _ string, _, _ []string) (*models.DocumentAnalysis, error) {
			return nil, errors.New("analysis failed")
		},
	}

	ctx := setupAnalyzeRegistry(t, analyzer, topicStore, tagStore)
	stage := newAnalyzeStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"},
		Job:     &models.Job{ID: "job-1"},
		Content: "test content",
	}

	// Enrich swallows errors and returns dc unchanged.
	result, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("Enrich should not propagate errors, got: %v", err)
	}
	if result.Summary != "" {
		t.Errorf("Summary should be empty on analyzer failure, got %q", result.Summary)
	}
}

func TestAnalyzeStage_EmptyVocabularies(t *testing.T) {
	var receivedTopics, receivedTags []string
	analyzer := &mockAnalyzer{
		OnAnalyze: func(_ context.Context, _ string, topics, tags []string) (*models.DocumentAnalysis, error) {
			receivedTopics = topics
			receivedTags = tags
			return &models.DocumentAnalysis{
				Summary:  "summary",
				Language: "en",
			}, nil
		},
	}
	topicStore := &mockIngestTopics{
		OnListTopicsByTenant: func(_ context.Context, _ string) ([]*models.Topic, error) {
			return nil, nil
		},
	}
	tagStore := &mockIngestTags{
		OnListTagsByTenant: func(_ context.Context, _ string) ([]*models.Tag, error) {
			return nil, nil
		},
	}

	ctx := setupAnalyzeRegistry(t, analyzer, topicStore, tagStore)
	stage := newAnalyzeStage()

	dc := &DocumentContext{
		Version: &models.DocumentVersion{ID: "ver-1", DocumentID: "doc-1", TenantID: "tenant-1"},
		Job:     &models.Job{ID: "job-1"},
		Content: "test content",
	}

	_, err := stage.Process(ctx, dc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receivedTopics) != 0 {
		t.Errorf("expected empty topics, got %v", receivedTopics)
	}
	if len(receivedTags) != 0 {
		t.Errorf("expected empty tags, got %v", receivedTags)
	}
}
