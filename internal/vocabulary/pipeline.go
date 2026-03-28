package vocabulary

import (
	"context"

	"github.com/zoobz-io/pipz"
)

// PipelineID is the stage identity for the vocabulary validation pipeline.
var PipelineID = pipz.NewIdentity("vocabulary-pipeline", "Vocabulary validation and classification pipeline")

// Pipeline orchestrates vocabulary validation and injection classification.
type Pipeline struct {
	sequence *pipz.Sequence[*Context]
}

// New creates a new vocabulary pipeline.
func New() *Pipeline {
	seq := pipz.NewSequence(
		PipelineID,
		newValidateStage(),
		newClassifyStage(),
	)
	return &Pipeline{sequence: seq}
}

// Process runs the vocabulary through validation and classification.
// Returns an error if validation fails or content is flagged as injection.
func (p *Pipeline) Process(ctx context.Context, tenantID, name, description string) error {
	vc := &Context{
		TenantID:    tenantID,
		Name:        name,
		Description: description,
	}
	_, err := p.sequence.Process(ctx, vc)
	return err
}

// ProcessUpdate runs the vocabulary through validation and classification for updates.
func (p *Pipeline) ProcessUpdate(ctx context.Context, id, name, description string) error {
	vc := &Context{
		ID:          id,
		Name:        name,
		Description: description,
	}
	_, err := p.sequence.Process(ctx, vc)
	return err
}
