//go:build testing

package stores

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/zoobz-io/grub"
	"github.com/zoobz-io/lucene"
)

// mockSearchProvider is a test mock for grub.SearchProvider.
type mockSearchProvider struct {
	OnIndex      func(ctx context.Context, index, id string, doc []byte) error
	OnIndexBatch func(ctx context.Context, index string, docs map[string][]byte) error
	OnGet        func(ctx context.Context, index, id string) ([]byte, error)
	OnDelete     func(ctx context.Context, index, id string) error
	OnSearch     func(ctx context.Context, index string, search *lucene.Search) (*grub.SearchResponse, error)
}

func (m *mockSearchProvider) Index(ctx context.Context, index, id string, doc []byte) error {
	if m.OnIndex != nil {
		return m.OnIndex(ctx, index, id, doc)
	}
	return nil
}
func (m *mockSearchProvider) IndexBatch(ctx context.Context, index string, docs map[string][]byte) error {
	if m.OnIndexBatch != nil {
		return m.OnIndexBatch(ctx, index, docs)
	}
	return nil
}
func (m *mockSearchProvider) Get(ctx context.Context, index, id string) ([]byte, error) {
	if m.OnGet != nil {
		return m.OnGet(ctx, index, id)
	}
	return nil, nil
}
func (m *mockSearchProvider) Delete(ctx context.Context, index, id string) error {
	if m.OnDelete != nil {
		return m.OnDelete(ctx, index, id)
	}
	return nil
}
func (m *mockSearchProvider) DeleteBatch(_ context.Context, _ string, _ []string) error {
	return nil
}
func (m *mockSearchProvider) Exists(_ context.Context, _, _ string) (bool, error) {
	return false, nil
}
func (m *mockSearchProvider) Search(ctx context.Context, index string, search *lucene.Search) (*grub.SearchResponse, error) {
	if m.OnSearch != nil {
		return m.OnSearch(ctx, index, search)
	}
	return &grub.SearchResponse{}, nil
}
func (m *mockSearchProvider) Count(_ context.Context, _ string, _ lucene.Query) (int64, error) {
	return 0, nil
}
func (m *mockSearchProvider) Refresh(_ context.Context, _ string) error {
	return nil
}

// jsonBytes marshals v to JSON bytes for mock search results.
func jsonBytes(t *testing.T, v any) []byte {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshaling: %v", err)
	}
	return data
}
