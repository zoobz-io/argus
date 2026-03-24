# handlers

HTTP handlers for the admin API surface.

## Purpose

Define HTTP endpoints for internal team operations. Admin handlers provide system-wide visibility, bulk operations, and administrative capabilities not exposed to customers.

## Pattern

```go
// admin/handlers/providers.go
package handlers

import (
    "github.com/zoobzio/rocco"
    "github.com/zoobzio/sum"
    "github.com/zoobzio/argus/admin/contracts"
    "github.com/zoobzio/argus/admin/transformers"
    "github.com/zoobzio/argus/admin/wire"
)

var listAllProviders = rocco.GET[rocco.NoBody, wire.AdminProviderListResponse]("/providers", func(r *rocco.Request[rocco.NoBody]) (wire.AdminProviderListResponse, error) {
    store := sum.MustUse[contracts.Providers](r)
    page := cursorPageFromQuery(r.Params)
    result, err := store.ListProviders(r, page)
    if err != nil {
        return wire.AdminProviderListResponse{}, err
    }
    return transformers.ProvidersToAdminList(result, page.PageSize()), nil
}).
    WithSummary("List all providers").
    WithTags("providers").
    WithQueryParams("cursor", "limit").
    WithAuthentication()

var deleteAdminProvider = rocco.DELETE[rocco.NoBody, rocco.NoBody]("/providers/{id}", func(r *rocco.Request[rocco.NoBody]) (rocco.NoBody, error) {
    id, err := pathID(r.Params, "id")
    if err != nil {
        return rocco.NoBody{}, rocco.ErrBadRequest.WithMessage("invalid id")
    }
    store := sum.MustUse[contracts.Providers](r)
    if err := store.DeleteProvider(r, id); err != nil {
        return rocco.NoBody{}, ErrProviderNotFound
    }
    return rocco.NoBody{}, nil
}).
    WithSummary("Delete provider").
    WithTags("providers").
    WithPathParams("id").
    WithSuccessStatus(204).
    WithAuthentication().
    WithErrors(ErrProviderNotFound)
```

## Handler Registration

```go
// admin/handlers/handlers.go
package handlers

import "github.com/zoobzio/rocco"

func All() []rocco.Endpoint {
    return []rocco.Endpoint{
        ListUsers,
        GetUser,
        SearchUsers,
        ImpersonateUser,
        // ... all admin handlers
    }
}
```

## Error Definitions

```go
// admin/handlers/errors.go
package handlers

import "github.com/zoobzio/rocco"

var (
    ErrUserNotFound    = rocco.ErrNotFound.WithMessage("user not found")
    ErrCannotImpersonate = rocco.ErrForbidden.WithMessage("cannot impersonate this user")
)
```

## Guidelines

- Admin handlers expose system-wide, cross-tenant operations
- Include list endpoints with cursor-based pagination
- Include delete operations not exposed in the public API
- Expose more data than public handlers (no masking)
- **Handlers are thin dispatchers** — parse params, call ONE contract method, transform result, return
- Never construct models or build queries in handlers — that logic lives in stores
- Register all handlers in the `All()` function
