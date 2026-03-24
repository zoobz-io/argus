# stores

Shared data access layer implementations.

## Purpose

Implement storage operations using `sum.Database`, `sum.Store`, or `sum.Bucket` wrappers. Stores are shared across all API surfaces — the same store can satisfy both public and admin contracts.

A single `Users` store might satisfy:
- `api/contracts.Users` (Get, Set, GetByLogin)
- `admin/contracts.Users` (Get, Set, Delete, List, Count)

The store implements all methods; each contract exposes only what that surface needs.

## Pattern

Stores expose **domain methods** that encapsulate all business logic — model construction, fetch-then-update, query building, pagination. The embedded `*sum.Database` provides low-level access (`Select`, `Query`, `Set`, `Delete`) but these are only used internally by the store's methods, never exposed through contracts.

```go
// stores/providers.go
package stores

import (
    "context"
    "fmt"

    "github.com/jmoiron/sqlx"
    "github.com/zoobzio/astql"
    "github.com/zoobz-io/sum"
    "github.com/zoobzio/argus/models"
)

type Providers struct {
    *sum.Database[models.Provider]
}

func NewProviders(db *sqlx.DB, renderer astql.Renderer) (*Providers, error) {
    database, err := sum.NewDatabase[models.Provider](db, "providers", renderer)
    if err != nil {
        return nil, err
    }
    return &Providers{Database: database}, nil
}

// GetProvider retrieves a provider by ID.
func (s *Providers) GetProvider(ctx context.Context, id int64) (*models.Provider, error) {
    return s.Select().
        Where("id", "=", ":id").
        Exec(ctx, map[string]any{"id": id})
}

// CreateProvider constructs and persists a new provider.
func (s *Providers) CreateProvider(ctx context.Context, tenantID int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error) {
    p := &models.Provider{TenantID: tenantID, Type: providerType, Name: name, Credentials: credentials, Active: true}
    if err := s.Set(ctx, "", p); err != nil {
        return nil, fmt.Errorf("creating provider: %w", err)
    }
    return p, nil
}

// UpdateProvider fetches, mutates, and persists a provider.
func (s *Providers) UpdateProvider(ctx context.Context, id int64, providerType models.ProviderType, name string, credentials string) (*models.Provider, error) {
    p, err := s.GetProvider(ctx, id)
    if err != nil {
        return nil, err
    }
    p.Type = providerType
    p.Name = name
    p.Credentials = credentials
    if err := s.Set(ctx, fmt.Sprintf("%d", id), p); err != nil {
        return nil, fmt.Errorf("updating provider: %w", err)
    }
    return p, nil
}

// ListProvidersByTenant returns providers for a tenant using cursor pagination.
func (s *Providers) ListProvidersByTenant(ctx context.Context, tenantID int64, page models.CursorPage) (*models.CursorResult[models.Provider], error) {
    limit := page.PageSize()
    q := s.Query().Where("tenant_id", "=", ":tenant_id").OrderBy("id", "ASC").Limit(limit + 1)
    params := map[string]any{"tenant_id": tenantID}
    if page.Cursor != nil {
        q = q.Where("id", ">", ":cursor")
        params["cursor"] = *page.Cursor
    }
    items, err := q.Exec(ctx, params)
    if err != nil {
        return nil, err
    }
    return cursorResult(items, limit), nil
}
```

## Aggregate Factory

```go
// stores/stores.go
package stores

import (
    "github.com/jmoiron/sqlx"
    "github.com/zoobzio/astql"
    "github.com/zoobzio/grub"
)

type Stores struct {
    Users        *Users
    Repositories *Repositories
    // ... all stores
}

func New(db *sqlx.DB, renderer astql.Renderer, bucket grub.BucketProvider) (*Stores, error) {
    users, err := NewUsers(db, renderer)
    if err != nil {
        return nil, err
    }
    // ... create all stores

    return &Stores{
        Users: users,
        // ...
    }, nil
}
```

## Multi-Surface Registration

The same store instance can be registered against different contracts in each binary:

```go
// cmd/app/main.go (public API)
allStores, _ := stores.New(db, renderer, bucket)
sum.Register[apicontracts.Users](k, allStores.Users)  // Minimal interface

// cmd/admin/main.go (admin API)
allStores, _ := stores.New(db, renderer, bucket)
sum.Register[admincontracts.Users](k, allStores.Users)  // Richer interface
```

## Guidelines

- One store per domain entity
- Embed `*sum.Database[Model]` for SQL stores
- Use `sum.NewStore[Model]` for key-value stores
- Use `sum.NewBucket[Model]` for object storage
- **Stores hold ALL business logic** — model construction, fetch-then-update, query building, pagination
- Use the Soy query builder for type-safe queries (never raw SQL)
- Expose domain methods (`CreateProvider`, `ListProvidersByTenant`), NOT raw storage methods (`Get`, `Set`)
- Keep `stores.go` as the aggregate factory
- Stores are shared — implement all methods any surface might need
- Register stores against surface-specific contracts in each binary's main.go
