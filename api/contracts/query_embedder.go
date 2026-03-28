package contracts

import (
	"context"

	"github.com/zoobz-io/vex"
)

// QueryEmbedder defines the contract for generating query embeddings.
type QueryEmbedder interface {
	// EmbedQuery generates a vector embedding optimized for search queries.
	EmbedQuery(ctx context.Context, text string) (vex.Vector, error)
}
