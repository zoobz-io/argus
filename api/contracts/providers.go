package contracts

import (
	"context"

	"github.com/zoobz-io/argus/models"
)

// Providers defines the contract for provider data access.
type Providers interface {
	// Get retrieves a provider by primary key.
	Get(ctx context.Context, key string) (*models.Provider, error)
	// Set creates or updates a provider.
	Set(ctx context.Context, key string, provider *models.Provider) error
}
