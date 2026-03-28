package ingest

import (
	"context"
	"fmt"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/models"
)

func newIndexStage() pipz.Chainable[*DocumentContext] {
	return pipz.Apply(
		IndexID,
		func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
			search := sum.MustUse[intcontracts.IngestSearch](ctx)
			jobs := sum.MustUse[intcontracts.IngestJobs](ctx)

			idx := &models.DocumentVersionIndex{
				VersionID:    dc.Version.ID,
				DocumentID:   dc.Version.DocumentID,
				TenantID:     dc.Version.TenantID,
				ProviderID:   dc.Document.ProviderID,
				DocumentName: dc.Document.Name,
				MimeType:     dc.Document.MimeType,
				Content:      dc.Content,
				Summary:      dc.Summary,
				Language:     dc.Language,
				Topics:       dc.Topics,
				Tags:         dc.Tags,
				Embedding:    dc.Embedding,
				CreatedAt:    dc.Version.CreatedAt,
				UpdatedAt:    dc.Document.UpdatedAt,
			}

			if err := search.IndexVersion(ctx, idx); err != nil {
				return dc, fmt.Errorf("indexing version: %w", err)
			}

			if err := jobs.UpdateJobStatus(ctx, dc.Job.ID, models.JobCompleted, nil); err != nil {
				return dc, fmt.Errorf("updating job status: %w", err)
			}

			capitan.Info(ctx, events.IngestIndexed,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
			)
			return dc, nil
		},
	)
}
