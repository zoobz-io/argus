package ingest

import (
	"context"
	"fmt"
	"strings"

	"github.com/zoobz-io/zyn"

	"github.com/zoobz-io/argus/models"
)

// extractionSynapse abstracts the zyn synapse so it can be stubbed in tests.
type extractionSynapse interface {
	FireWithInput(ctx context.Context, session *zyn.Session, input zyn.ExtractionInput) (models.DocumentAnalysis, error)
}

// SynapseAnalyzer adapts a zyn.ExtractionSynapse to satisfy the Analyzer contract.
type SynapseAnalyzer struct {
	synapse extractionSynapse
}

// NewSynapseAnalyzer creates an Analyzer backed by a zyn extraction synapse.
func NewSynapseAnalyzer(synapse *zyn.ExtractionSynapse[models.DocumentAnalysis]) *SynapseAnalyzer {
	return &SynapseAnalyzer{synapse: synapse}
}

// Analyze extracts structured metadata from document content, injecting the
// tenant's vocabulary into the prompt context.
func (a *SynapseAnalyzer) Analyze(ctx context.Context, content string, topics, tags []string) (*models.DocumentAnalysis, error) {
	var contextParts []string

	if len(topics) > 0 {
		contextParts = append(contextParts,
			fmt.Sprintf("Allowed topics (select ONLY from this list): %s", strings.Join(topics, ", ")))
	} else {
		contextParts = append(contextParts, "No topics are defined. Return an empty topics array.")
	}

	if len(tags) > 0 {
		contextParts = append(contextParts,
			fmt.Sprintf("Allowed tags (select ONLY from this list): %s", strings.Join(tags, ", ")))
	} else {
		contextParts = append(contextParts, "No tags are defined. Return an empty tags array.")
	}

	input := zyn.ExtractionInput{
		Text:    content,
		Context: strings.Join(contextParts, "\n"),
	}

	result, err := a.synapse.FireWithInput(ctx, zyn.NewSession(), input)
	if err != nil {
		return nil, fmt.Errorf("analyzing document: %w", err)
	}

	return &result, nil
}
