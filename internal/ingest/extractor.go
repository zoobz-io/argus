package ingest

import (
	"context"

	"github.com/zoobz-io/zyn"
)

// SynapseExtractor adapts a zyn.TransformSynapse to satisfy the TextExtractor contract.
type SynapseExtractor struct {
	synapse *zyn.TransformSynapse
}

// NewSynapseExtractor creates a TextExtractor backed by a zyn transform synapse.
func NewSynapseExtractor(synapse *zyn.TransformSynapse) *SynapseExtractor {
	return &SynapseExtractor{synapse: synapse}
}

// Extract cleans raw pre-processed document text into readable content.
func (e *SynapseExtractor) Extract(ctx context.Context, rawText string) (string, error) {
	session := zyn.NewSession()
	return e.synapse.Fire(ctx, session, rawText)
}
