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
			versions := sum.MustUse[intcontracts.IngestVersions](ctx)

			idx := &models.DocumentVersionIndex{
				VersionID:    dc.Version.ID,
				DocumentID:   dc.Version.DocumentID,
				TenantID:     dc.Version.TenantID,
				ProviderID:   dc.Document.ProviderID,
				DocumentName: dc.Document.Name,
				MimeType:     dc.Document.MimeType,
				Content:      dc.Content,
				Summary:      dc.Summary,
				Embedding:    dc.Embedding,
			}

			if err := search.IndexVersion(ctx, idx); err != nil {
				return dc, fmt.Errorf("indexing version: %w", err)
			}

			if err := versions.UpdateExtractionStatus(ctx, dc.Version.ID, models.ExtractionCompleted); err != nil {
				return dc, fmt.Errorf("updating extraction status: %w", err)
			}

			capitan.Info(ctx, events.IngestIndexed,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
			)
			return dc, nil
		},
	)
}
