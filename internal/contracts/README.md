# contracts

Interface definitions for internal services.

## Purpose

Define boundaries between internal consumers and their dependencies. Internal contracts serve the same role as `api/contracts/` and `admin/contracts/` but for services that are not exposed through any API surface — pipelines, background workers, and internal orchestration.

Consumers resolve implementations at runtime via `sum.MustUse[Contract](ctx)`. Implementations are registered in `cmd/*/main.go` before `sum.Freeze(k)`.

## Pattern

```go
// internal/contracts/ingest.go
package contracts

import "context"

type Summarizer interface {
    // Summarize generates a text summary of the given content.
    Summarize(ctx context.Context, content string) (string, error)
}
```

Registration in main.go:

```go
sum.Register[intcontracts.Summarizer](k, summarizer)
```

Resolution at call time:

```go
summarizer := sum.MustUse[intcontracts.Summarizer](ctx)
summary, err := summarizer.Summarize(ctx, content)
```

## Guidelines

- One interface per concern, not per implementation
- Methods are **domain operations** from the consumer's perspective
- First parameter is always `context.Context`
- Return `error` as the last return value
- Keep interfaces minimal — only what the consumer actually needs
- Adapters (e.g., wrapping a zyn synapse to satisfy `Summarizer`) live in the consuming package, not here
