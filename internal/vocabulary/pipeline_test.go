//go:build testing

package vocabulary

import (
	"context"
	"strings"
	"testing"

	"github.com/zoobz-io/sum"
	"google.golang.org/grpc"

	"github.com/zoobz-io/argus/config"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// pipelineMockClassifier is a separate mock for pipeline tests to avoid
// redeclaration with classify_test.go (same package, same build tag).
type pipelineMockClassifier struct {
	OnClassifyText func(ctx context.Context, in *proto.ClassifyRequest) (*proto.ClassifyResponse, error)
}

func (m *pipelineMockClassifier) ClassifyText(ctx context.Context, in *proto.ClassifyRequest, _ ...grpc.CallOption) (*proto.ClassifyResponse, error) {
	return m.OnClassifyText(ctx, in)
}

func setupPipelineRegistry(t *testing.T, classifier intcontracts.Classifier) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)
	k := sum.Start()
	_ = sum.Config[config.Classify](context.Background(), k, nil)
	sum.Register[intcontracts.Classifier](k, classifier)
	sum.Freeze(k)
	return context.Background()
}

func safeClassifier() *pipelineMockClassifier {
	return &pipelineMockClassifier{
		OnClassifyText: func(_ context.Context, _ *proto.ClassifyRequest) (*proto.ClassifyResponse, error) {
			return &proto.ClassifyResponse{Safe: true, Score: 0.02, Label: "SAFE"}, nil
		},
	}
}

func unsafeClassifier() *pipelineMockClassifier {
	return &pipelineMockClassifier{
		OnClassifyText: func(_ context.Context, _ *proto.ClassifyRequest) (*proto.ClassifyResponse, error) {
			return &proto.ClassifyResponse{Safe: false, Score: 0.99, Label: "INJECTION"}, nil
		},
	}
}

func TestPipeline_Process_HappyPath(t *testing.T) {
	ctx := setupPipelineRegistry(t, safeClassifier())
	p := New()

	err := p.Process(ctx, "tenant-1", "Finance", "Financial topics and reporting.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipeline_Process_ValidationFailure(t *testing.T) {
	ctx := setupPipelineRegistry(t, safeClassifier())
	p := New()

	// Empty name should fail validation before reaching classify.
	err := p.Process(ctx, "tenant-1", "", "A valid description.")
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
	if !strings.Contains(err.Error(), "vocabulary validation failed") {
		t.Errorf("error should mention vocabulary validation failed, got %q", err.Error())
	}
}

func TestPipeline_Process_ClassificationRejection(t *testing.T) {
	ctx := setupPipelineRegistry(t, unsafeClassifier())
	p := New()

	err := p.Process(ctx, "tenant-1", "Finance", "Financial topics.")
	if err == nil {
		t.Fatal("expected error for rejected classification")
	}
	if !strings.Contains(err.Error(), "prompt injection") {
		t.Errorf("error should mention prompt injection, got %q", err.Error())
	}
}

func TestPipeline_ProcessUpdate_HappyPath(t *testing.T) {
	ctx := setupPipelineRegistry(t, safeClassifier())
	p := New()

	err := p.ProcessUpdate(ctx, "vocab-1", "Finance", "Updated financial topics.")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPipeline_ProcessUpdate_ValidationFailure(t *testing.T) {
	ctx := setupPipelineRegistry(t, safeClassifier())
	p := New()

	err := p.ProcessUpdate(ctx, "vocab-1", "", "A valid description.")
	if err == nil {
		t.Fatal("expected error for invalid name on update")
	}
	if !strings.Contains(err.Error(), "vocabulary validation failed") {
		t.Errorf("error should mention vocabulary validation failed, got %q", err.Error())
	}
}

func TestPipeline_ProcessUpdate_ClassificationRejection(t *testing.T) {
	ctx := setupPipelineRegistry(t, unsafeClassifier())
	p := New()

	err := p.ProcessUpdate(ctx, "vocab-1", "Finance", "Financial topics.")
	if err == nil {
		t.Fatal("expected error for rejected classification on update")
	}
	if !strings.Contains(err.Error(), "prompt injection") {
		t.Errorf("error should mention prompt injection, got %q", err.Error())
	}
}
