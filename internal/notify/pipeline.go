package notify

import (
	"context"

	"github.com/zoobz-io/pipz"
)

// Stage identities for pipeline introspection and resolution.
var (
	PipelineID = pipz.NewIdentity("notify-pipeline", "Notification fan-out pipeline")
	AssignID   = pipz.NewIdentity("notify-assign", "Assign notification fields from subscription")
	IndexID    = pipz.NewIdentity("notify-index", "Index notification in search")
	HintID     = pipz.NewIdentity("notify-hint", "Emit notification hint signal")
)

// Pipeline orchestrates notification fan-out through assignment, indexing,
// and hint emission stages.
type Pipeline struct {
	sequence *pipz.Sequence[*FanOutItem]
}

// New creates a new notification fan-out pipeline.
func New() *Pipeline {
	seq := pipz.NewSequence(
		PipelineID,
		newAssignStage(),
		newIndexStage(),
		newHintStage(),
	)
	return &Pipeline{sequence: seq}
}

// Process runs a fan-out item through the full notification pipeline.
func (p *Pipeline) Process(ctx context.Context, item *FanOutItem) (*FanOutItem, error) {
	return p.sequence.Process(ctx, item)
}
