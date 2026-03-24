# contracts

Interface definitions for the admin API surface.

## Purpose

Define the boundaries between layers for admin operations. Admin contracts may expose broader capabilities than public contracts — system-wide access, cross-tenant queries, delete operations, and internal data.

Like public contracts, admin contracts define **domain operations**, not raw storage primitives. The store holds all business logic.

## Pattern

```go
// admin/contracts/providers.go
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
    // DeleteProvider removes a provider.
    DeleteProvider(ctx context.Context, id int64) error
    // ListProviders retrieves a paginated list of all providers (cross-tenant).
    ListProviders(ctx context.Context, page models.CursorPage) (*models.CursorResult[models.Provider], error)
}
```

## Guidelines

- Admin contracts may expose more methods than public equivalents
- Include cross-tenant list operations with cursor pagination
- Include delete operations not exposed to the public API
- Methods are **domain operations** (e.g., `DeleteProvider`, `ListProviders`), NOT raw storage methods (`Get`, `Set`, `Delete`)
- Document each method with a brief comment
- Keep interfaces focused - prefer multiple small interfaces over one large one

## Naming

- Interface names are nouns: `Providers`, `Documents`, `Tenants`
- Method names describe the domain operation: `GetProvider`, `DeleteProvider`, `ListProviders`
- Use `ByX` suffix for scoped queries: `ListProvidersByTenant`
