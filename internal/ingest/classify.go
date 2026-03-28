package ingest

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

// ClassifyID is the stage identity for prompt injection classification.
var ClassifyID = pipz.NewIdentity("classify", "Classify content for prompt injection")

// InjectionThreshold is the score above which content is flagged as injection.
const InjectionThreshold = 0.5

func newClassifyStage() pipz.Chainable[*DocumentContext] {
	return pipz.Apply(ClassifyID, func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
		classifier := sum.MustUse[intcontracts.Classifier](ctx)

		resp, err := classifier.ClassifyText(ctx, &proto.ClassifyRequest{
			Text: dc.Content,
		})
		if err != nil {
			// Classification service unavailable — log and continue.
			// Failing open is a policy decision. Change to fail-closed
			// by returning the error instead.
			capitan.Warn(ctx, events.IngestClassifyUnavailable,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestErrorKey.Field(err),
			)
			return dc, nil
		}

		if !resp.Safe {
			capitan.Warn(ctx, events.IngestClassifyRejected,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
			)
			return dc, fmt.Errorf("content flagged as prompt injection (score=%.2f, label=%s)", resp.Score, resp.Label)
		}

		return dc, nil
	})
}
