// Package wire defines request and response types for the public API surface.
package wire

import "github.com/zoobz-io/sum"

// RegisterBoundaries registers all API wire type boundaries with the given key.
func RegisterBoundaries(k sum.Key) {
	sum.NewBoundary[TenantResponse](k)
	sum.NewBoundary[ProviderResponse](k)
	sum.NewBoundary[WatchedPathResponse](k)
	sum.NewBoundary[DocumentResponse](k)
	sum.NewBoundary[DocumentVersionResponse](k)
	sum.NewBoundary[SearchResponse](k)
	sum.NewBoundary[ProviderListResponse](k)
	sum.NewBoundary[WatchedPathListResponse](k)
	sum.NewBoundary[DocumentListResponse](k)
	sum.NewBoundary[DocumentVersionListResponse](k)
	sum.NewBoundary[TopicResponse](k)
	sum.NewBoundary[TopicListResponse](k)
	sum.NewBoundary[TagResponse](k)
	sum.NewBoundary[TagListResponse](k)
	sum.NewBoundary[UserResponse](k)
	sum.NewBoundary[UserListResponse](k)
	sum.NewBoundary[SubscriptionResponse](k)
	sum.NewBoundary[SubscriptionListResponse](k)
	sum.NewBoundary[NotificationResponse](k)
	sum.NewBoundary[NotificationListResponse](k)
}
