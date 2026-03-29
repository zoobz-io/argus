package notify

import (
	"context"
	"fmt"

	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

func newIndexStage() pipz.Chainable[*FanOutItem] {
	return pipz.Apply(
		IndexID,
		func(ctx context.Context, item *FanOutItem) (*FanOutItem, error) {
			indexer := sum.MustUse[intcontracts.NotifyIndexer](ctx)

			if err := indexer.Index(ctx, item.Notification); err != nil {
				return item, fmt.Errorf("indexing notification: %w", err)
			}
			return item, nil
		},
	)
}
