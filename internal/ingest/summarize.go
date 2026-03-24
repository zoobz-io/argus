package ingest

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

func newSummarizeStage() pipz.Chainable[*DocumentContext] {
	return pipz.Enrich(
		SummarizeID,
		func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
			summarizer := sum.MustUse[intcontracts.Summarizer](ctx)

			summary, err := summarizer.Summarize(ctx, dc.Content)
			if err != nil {
				capitan.Warn(ctx, events.IngestSummarizeFailed,
					events.IngestVersionIDKey.Field(dc.Version.ID),
					events.IngestErrorKey.Field(err),
				)
				return dc, err
			}
			dc.Summary = summary
			capitan.Info(ctx, events.IngestSummarized,
				events.IngestVersionIDKey.Field(dc.Version.ID),
			)
			return dc, nil
		},
	)
}
