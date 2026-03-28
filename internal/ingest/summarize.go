package ingest

import (
	"context"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

func newAnalyzeStage() pipz.Chainable[*DocumentContext] {
	return pipz.Enrich(
		AnalyzeID,
		func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
			analyzer := sum.MustUse[intcontracts.Analyzer](ctx)
			topicStore := sum.MustUse[intcontracts.IngestTopics](ctx)
			tagStore := sum.MustUse[intcontracts.IngestTags](ctx)

			// Load tenant vocabularies.
			topics, err := topicStore.ListTopicsByTenant(ctx, dc.Version.TenantID)
			if err != nil {
				return dc, err
			}
			tags, err := tagStore.ListTagsByTenant(ctx, dc.Version.TenantID)
			if err != nil {
				return dc, err
			}

			// Build name lists for the LLM prompt.
			topicNames := make([]string, len(topics))
			for i, t := range topics {
				topicNames[i] = t.Name
			}
			tagNames := make([]string, len(tags))
			for i, t := range tags {
				tagNames[i] = t.Name
			}

			analysis, err := analyzer.Analyze(ctx, dc.Content, topicNames, tagNames)
			if err != nil {
				capitan.Warn(ctx, events.IngestSummarizeFailed,
					events.IngestVersionIDKey.Field(dc.Version.ID),
					events.IngestErrorKey.Field(err),
				)
				return dc, err
			}

			dc.Summary = analysis.Summary
			dc.Language = analysis.Language
			dc.Topics = analysis.Topics
			dc.Tags = analysis.Tags

			capitan.Info(ctx, events.IngestSummarized,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
				events.IngestJobIDKey.Field(dc.Job.ID),
			)
			return dc, nil
		},
	)
}
