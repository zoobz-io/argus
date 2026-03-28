//go:build integration

package integration

import (
	"context"
	"math/rand"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/proto"
	"github.com/zoobz-io/vex"
	"google.golang.org/grpc"
)

// stubAnalyzer returns canned analysis results for pipeline integration tests.
type stubAnalyzer struct{}

func (s *stubAnalyzer) Analyze(_ context.Context, content string, topics, tags []string) (*models.DocumentAnalysis, error) {
	// Pick the first topic/tag if available.
	var matchedTopics, matchedTags []string
	if len(topics) > 0 {
		matchedTopics = topics[:1]
	}
	if len(tags) > 0 {
		matchedTags = tags[:1]
	}
	return &models.DocumentAnalysis{
		Summary:  "Test summary of: " + content[:min(50, len(content))],
		Language: "en",
		Topics:   matchedTopics,
		Tags:     matchedTags,
	}, nil
}

// stubEmbedder returns a fixed-dimension random vector for pipeline integration tests.
type stubEmbedder struct {
	dimensions int
}

func (s *stubEmbedder) Embed(_ context.Context, _ string) (vex.Vector, error) {
	vec := make(vex.Vector, s.dimensions)
	for i := range vec {
		vec[i] = rand.Float32()
	}
	return vec, nil
}

// stubOCR returns the raw bytes as text (only used if image extraction is triggered).
type stubOCR struct{}

func (s *stubOCR) ExtractText(_ context.Context, in *proto.ExtractTextRequest, _ ...grpc.CallOption) (*proto.ExtractTextResponse, error) {
	return &proto.ExtractTextResponse{
		Text:       string(in.Document),
		Confidence: 0.99,
	}, nil
}

// stubQueryEmbedder satisfies apicontracts.QueryEmbedder.
type stubQueryEmbedder struct {
	dimensions int
}

func (s *stubQueryEmbedder) EmbedQuery(_ context.Context, _ string) (vex.Vector, error) {
	vec := make(vex.Vector, s.dimensions)
	for i := range vec {
		vec[i] = rand.Float32()
	}
	return vec, nil
}

// stubClassifier always returns safe. For real injection testing, use the
// Python DeBERTa sidecar via testcontainer.
type stubClassifier struct{}

func (s *stubClassifier) ClassifyText(_ context.Context, in *proto.ClassifyRequest, _ ...grpc.CallOption) (*proto.ClassifyResponse, error) {
	return &proto.ClassifyResponse{
		Safe:  true,
		Score: 0.01,
		Label: "SAFE",
	}, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
