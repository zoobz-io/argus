package vocabulary

import (
	"context"
	"fmt"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// ClassifyID is the stage identity for vocabulary injection classification.
var ClassifyID = pipz.NewIdentity("vocabulary-classify", "Classify vocabulary for prompt injection")

func newClassifyStage() pipz.Chainable[*Context] {
	return pipz.Apply(ClassifyID, func(ctx context.Context, vc *Context) (*Context, error) {
		classifier := sum.MustUse[intcontracts.Classifier](ctx)

		// Classify the combined name + description as a single text block.
		text := vc.Name + ": " + vc.Description
		resp, err := classifier.ClassifyText(ctx, &proto.ClassifyRequest{
			Text: text,
		})
		if err != nil {
			// Fail closed for vocabulary — unlike document ingestion,
			// vocabulary is user-supplied and directly injected into prompts.
			capitan.Warn(ctx, events.VocabularyClassifyUnavailable,
				events.VocabularyNameKey.Field(vc.Name),
				events.IngestErrorKey.Field(err),
			)
			return vc, fmt.Errorf("injection classifier unavailable: %w", err)
		}

		if !resp.Safe {
			capitan.Warn(ctx, events.VocabularyClassifyRejected,
				events.VocabularyNameKey.Field(vc.Name),
			)
			return vc, fmt.Errorf("vocabulary rejected: potential prompt injection (score=%.2f)", resp.Score)
		}

		return vc, nil
	})
}
