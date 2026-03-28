//go:build testing

package wire

import (
	"context"
	"errors"
	"testing"

	argustest "github.com/zoobz-io/argus/testing"
	"github.com/zoobz-io/sum"
)

// setupRegistry registers all wire and model boundaries via argustest.
func setupRegistry(t *testing.T) context.Context {
	t.Helper()
	return argustest.SetupRegistry(t, argustest.WithBoundaries(RegisterBoundaries))
}

// failBoundary is a mock sum.Boundary that always returns an error from Send.
type failBoundary[T any] struct {
	err error
}

func (f *failBoundary[T]) Send(_ context.Context, _ T) (T, error) {
	var zero T
	return zero, f.err
}

func (f *failBoundary[T]) Receive(_ context.Context, obj T) (T, error) { return obj, nil }
func (f *failBoundary[T]) Store(_ context.Context, obj T) (T, error)   { return obj, nil }
func (f *failBoundary[T]) Load(_ context.Context, obj T) (T, error)    { return obj, nil }

// setupFailRegistry registers failing boundaries for all wire types.
func setupFailRegistry(t *testing.T) context.Context {
	t.Helper()
	sum.Reset()
	sum.New()
	k := sum.Start()

	sendErr := errors.New("send failed")
	sum.Register[sum.Boundary[TenantResponse]](k, &failBoundary[TenantResponse]{err: sendErr})
	sum.Register[sum.Boundary[ProviderResponse]](k, &failBoundary[ProviderResponse]{err: sendErr})
	sum.Register[sum.Boundary[WatchedPathResponse]](k, &failBoundary[WatchedPathResponse]{err: sendErr})
	sum.Register[sum.Boundary[DocumentResponse]](k, &failBoundary[DocumentResponse]{err: sendErr})
	sum.Register[sum.Boundary[DocumentVersionResponse]](k, &failBoundary[DocumentVersionResponse]{err: sendErr})
	sum.Register[sum.Boundary[SearchResponse]](k, &failBoundary[SearchResponse]{err: sendErr})
	sum.Register[sum.Boundary[ProviderListResponse]](k, &failBoundary[ProviderListResponse]{err: sendErr})
	sum.Register[sum.Boundary[WatchedPathListResponse]](k, &failBoundary[WatchedPathListResponse]{err: sendErr})
	sum.Register[sum.Boundary[DocumentListResponse]](k, &failBoundary[DocumentListResponse]{err: sendErr})
	sum.Register[sum.Boundary[DocumentVersionListResponse]](k, &failBoundary[DocumentVersionListResponse]{err: sendErr})
	sum.Register[sum.Boundary[TopicResponse]](k, &failBoundary[TopicResponse]{err: sendErr})
	sum.Register[sum.Boundary[TopicListResponse]](k, &failBoundary[TopicListResponse]{err: sendErr})
	sum.Register[sum.Boundary[TagResponse]](k, &failBoundary[TagResponse]{err: sendErr})
	sum.Register[sum.Boundary[TagListResponse]](k, &failBoundary[TagListResponse]{err: sendErr})

	sum.Freeze(k)
	t.Cleanup(sum.Reset)
	return context.Background()
}
