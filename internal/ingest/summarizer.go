package ingest

import (
	"context"

	"github.com/zoobz-io/zyn"
)

// SynapseSummarizer adapts a zyn.TransformSynapse to satisfy the Summarizer contract.
type SynapseSummarizer struct {
	synapse *zyn.TransformSynapse
}

// NewSynapseSummarizer creates a Summarizer backed by a zyn transform synapse.
func NewSynapseSummarizer(synapse *zyn.TransformSynapse) *SynapseSummarizer {
	return &SynapseSummarizer{synapse: synapse}
}

// Summarize generates a text summary using the underlying synapse.
func (s *SynapseSummarizer) Summarize(ctx context.Context, content string) (string, error) {
	session := zyn.NewSession()
	return s.synapse.Fire(ctx, session, content)
}
