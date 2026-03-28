package contracts

import "context"

// Vocabulary defines the vocabulary validation and classification pipeline.
type Vocabulary interface {
	// Process validates and classifies a vocabulary entry before creation.
	Process(ctx context.Context, tenantID, name, description string) error
	// ProcessUpdate validates and classifies a vocabulary entry before update.
	ProcessUpdate(ctx context.Context, id, name, description string) error
}
