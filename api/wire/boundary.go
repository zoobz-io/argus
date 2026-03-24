// Package wire defines request and response types for the public API surface.
package wire

import "github.com/zoobz-io/sum"

// RegisterBoundaries registers all API wire type boundaries with the given key.
func RegisterBoundaries(k sum.Key) error {
	if _, err := sum.NewBoundary[TenantResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[ProviderResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[WatchedPathResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[DocumentResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[DocumentVersionResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[SearchResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[ProviderListResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[WatchedPathListResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[DocumentListResponse](k); err != nil {
		return err
	}
	if _, err := sum.NewBoundary[DocumentVersionListResponse](k); err != nil {
		return err
	}
	return nil
}
