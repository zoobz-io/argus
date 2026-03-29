//go:build testing

package argustest

import (
	"time"

	"github.com/zoobz-io/argus/models"
)

var (
	FixtureTime  = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
	FixtureTime2 = time.Date(2026, 3, 16, 12, 0, 0, 0, time.UTC)
)

func NewTenant() *models.Tenant {
	return &models.Tenant{
		ID: "t1", Name: "Acme", Slug: "acme",
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewProvider() *models.Provider {
	return &models.Provider{
		ID: "p1", TenantID: "t1", Type: models.ProviderGoogleDrive,
		Name: "GDrive", Active: true, Credentials: "secret",
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewDocument() *models.Document {
	vid := "v1"
	return &models.Document{
		ID: "d1", TenantID: "t1", ProviderID: "p1", WatchedPathID: "wp1",
		ExternalID: "ext1", Name: "report.pdf", MimeType: "application/pdf",
		ObjectKey: "obj1", CurrentVersionID: &vid,
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewDocumentVersion() *models.DocumentVersion {
	return &models.DocumentVersion{
		ID: "v1", DocumentID: "d1", TenantID: "t1",
		VersionNumber: 3, ContentHash: "abc123",
		CreatedAt: FixtureTime,
	}
}

func NewWatchedPath() *models.WatchedPath {
	return &models.WatchedPath{
		ID: "wp1", TenantID: "t1", ProviderID: "p1",
		Path: "/docs", Active: true,
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewTopic() *models.Topic {
	return &models.Topic{
		ID: "tp1", TenantID: "t1",
		Name: "Security", Description: "Sec docs",
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewUser() *models.User {
	seen := FixtureTime2
	return &models.User{
		ID: "u1", ExternalID: "ext-u1", TenantID: "t1",
		Email: "user@example.com", DisplayName: "Jane Doe",
		Role: models.UserRoleViewer, Status: models.UserStatusActive,
		LastSeenAt: &seen,
		CreatedAt:  FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewTag() *models.Tag {
	return &models.Tag{
		ID: "tg1", TenantID: "t1",
		Name: "compliance", Description: "Compliance docs",
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewSubscription() *models.Subscription {
	return &models.Subscription{
		ID: "s1", UserID: "u1", TenantID: "t1",
		EventType: "ingest.completed", Channel: models.SubscriptionChannelInbox,
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewNotification() *models.Notification {
	return &models.Notification{
		ID: "n1", UserID: "u1", TenantID: "t1",
		Type: models.NotificationIngestCompleted, Status: models.NotificationUnread,
		DocumentID: "d1", VersionID: "v1", Message: "Ingest completed",
		CreatedAt: FixtureTime,
	}
}

func NewHook() *models.Hook {
	return &models.Hook{
		ID: "h1", TenantID: "t1", UserID: "u1",
		URL: "https://example.com/webhook", Secret: "whsec_test123",
		Active: true,
		CreatedAt: FixtureTime, UpdatedAt: FixtureTime2,
	}
}

func NewDelivery() *models.Delivery {
	return &models.Delivery{
		ID: "dl1", HookID: "h1", EventID: "evt1", TenantID: "t1",
		StatusCode: 200, Attempt: 1,
		CreatedAt: FixtureTime,
	}
}
