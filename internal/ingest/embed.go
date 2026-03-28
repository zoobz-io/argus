package ingest

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

func newEmbedStage() pipz.Chainable[*DocumentContext] {
	return pipz.Enrich(
		EmbedID,
		func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
			embedder := sum.MustUse[intcontracts.Embedder](ctx)

			vec, err := embedder.Embed(ctx, dc.Content)
			if err != nil {
				capitan.Warn(ctx, events.IngestEmbedFailed,
					events.IngestVersionIDKey.Field(dc.Version.ID),
					events.IngestJobIDKey.Field(dc.Job.ID),
					events.IngestErrorKey.Field(err),
				)
				return dc, err
			}
			dc.Embedding = vec
			capitan.Info(ctx, events.IngestEmbedded,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
				events.IngestJobIDKey.Field(dc.Job.ID),
			)
			return dc, nil
		},
	)
}
