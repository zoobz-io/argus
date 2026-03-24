# contracts

Interface definitions for the public API surface.

## Purpose

Define the boundaries between layers. Handlers depend on contracts, stores implement contracts. This enables testing with mocks and keeps dependencies flowing in one direction.

Contracts define **domain operations** — what the handler needs to accomplish, not raw storage primitives. The store holds all business logic; the contract is the seam between handler and store.

## Pattern

```go
// contracts/providers.go
package contracts

import (
    "context"

    "github.com/zoobzio/argus/models"
)

type Providers interface {
    // GetProvider retrieves a provider by ID.
    GetProvider(ctx context.Context, id int64) (*models.Provider, error)
    // CreateProvider creates a new provider for a tenant.
    CreateProvider(ctx context.Context, tenantID int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
    // UpdateProvider updates an existing provider.
    UpdateProvider(ctx context.Context, id int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error)
    // ListProvidersByTenant retrieves providers for a tenant using cursor pagination.
    ListProvidersByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Provider], error)
}
```

## Guidelines

- One interface per domain entity
- First parameter is always `context.Context`
- Return pointers for single entities, `*models.CursorResult[T]` for paginated lists
- Return `error` as the last return value
- Document each method with a brief comment
- Keep interfaces focused - prefer multiple small interfaces over one large one
- Methods are **domain operations** (e.g., `CreateProvider`, `ListProvidersByTenant`), NOT raw storage methods (`Get`, `Set`, `Delete`)
- Handlers should only need to call a single contract method per request

## Naming

- Interface names are nouns: `Providers`, `Documents`, `Tenants`
- Method names describe the domain operation: `GetProvider`, `CreateProvider`, `ListProvidersByTenant`
- Use `ByX` suffix for scoped queries: `ListProvidersByTenant`, `ListDocumentsByProvider`
