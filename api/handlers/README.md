# handlers

HTTP handlers for the public API surface.

## Purpose

Define HTTP endpoints using rocco's handler pattern. Handlers are the entry point for HTTP requests and orchestrate calls to contracts.

## Pattern

```go
// handlers/providers.go
package handlers

import (
    "github.com/zoobzio/rocco"
    "github.com/zoobzio/sum"
    "github.com/zoobzio/argus/api/contracts"
    "github.com/zoobzio/argus/api/transformers"
    "github.com/zoobzio/argus/api/wire"
)

var getProvider = rocco.GET[rocco.NoBody, wire.ProviderResponse]("/providers/{id}", func(r *rocco.Request[rocco.NoBody]) (wire.ProviderResponse, error) {
    id, err := pathID(r.Params, "id")
    if err != nil {
        return wire.ProviderResponse{}, rocco.ErrBadRequest.WithMessage("invalid id")
    }
    store := sum.MustUse[contracts.Providers](r)
    provider, err := store.GetProvider(r, id)
    if err != nil {
        return wire.ProviderResponse{}, ErrProviderNotFound
    }
    return transformers.ProviderToResponse(provider), nil
}).
    WithSummary("Get provider").
    WithTags("providers").
    WithPathParams("id").
    WithAuthentication().
    WithErrors(ErrProviderNotFound)

var createProvider = rocco.POST[wire.ProviderCreateRequest, wire.ProviderResponse]("/providers", func(r *rocco.Request[wire.ProviderCreateRequest]) (wire.ProviderResponse, error) {
    tid, err := tenantID(r.Identity)
    if err != nil {
        return wire.ProviderResponse{}, rocco.ErrBadRequest.WithMessage("invalid tenant")
    }
    store := sum.MustUse[contracts.Providers](r)
    provider, err := store.CreateProvider(r, tid, r.Body.Type, r.Body.Name, r.Body.Credentials)
    if err != nil {
        return wire.ProviderResponse{}, err
    }
    return transformers.ProviderToResponse(provider), nil
}).
    WithSummary("Create provider").
    WithTags("providers").
    WithSuccessStatus(201).
    WithAuthentication().
    WithErrors(rocco.ErrValidationFailed)
```

## Handler Registration

```go
// handlers/handlers.go
package handlers

import "github.com/zoobzio/rocco"

func All() []rocco.Endpoint {
    return []rocco.Endpoint{
        GetMe,
        UpdateMe,
        // ... all handlers
    }
}
```

## Error Definitions

```go
// handlers/errors.go
package handlers

import "github.com/zoobzio/rocco"

var (
    ErrUserNotFound = rocco.ErrNotFound.WithMessage("user not found")
    ErrMissingQuery = rocco.ErrBadRequest.WithMessage("query parameter required")
)
```

## Chainable Methods

| Method | Purpose |
|--------|---------|
| `.WithSummary()` | OpenAPI summary |
| `.WithDescription()` | OpenAPI description |
| `.WithTags()` | OpenAPI tag grouping |
| `.WithAuthentication()` | Require auth |
| `.WithPathParams()` | Define path variables |
| `.WithQueryParams()` | Document query parameters |
| `.WithErrors()` | Document expected errors |
| `.WithSuccessStatus()` | Override default 200 |

## Streaming (SSE)

For real-time updates, use `rocco.NewStreamHandler`:

```go
var StreamProgress = rocco.NewStreamHandler[rocco.NoBody, wire.ProgressUpdate](
    "progress-stream",
    http.MethodGet,
    "/jobs/{id}/progress",
    func(req *rocco.Request[rocco.NoBody], stream rocco.Stream[wire.ProgressUpdate]) error {
        for {
            select {
            case <-stream.Done():
                return nil // Client disconnected
            case update := <-progress:
                if err := stream.Send(update); err != nil {
                    return err
                }
            }
        }
    },
).WithPathParams("id").
    WithSummary("Stream job progress").
    WithAuthentication()
```

## Guidelines

- Handlers are module-level variables, not methods
- Use `rocco.NoBody` for requests without a body
- Retrieve contracts via `sum.MustUse[contracts.T](r)` (r is `*rocco.Request` which embeds context)
- Use transformers for model -> wire conversion
- **Handlers are thin dispatchers** — parse params, call ONE contract method, transform result, return
- Never construct models or build queries in handlers — that logic lives in stores
- Define domain-specific errors in `errors.go`
- Register all handlers in the `All()` function
- For streams: always check `stream.Done()` for client disconnect
