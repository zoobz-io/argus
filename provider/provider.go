// Package provider defines the contract that all cloud storage providers satisfy.
//
// The provider interface abstracts authentication, file listing, change detection,
// and content retrieval across cloud storage platforms (Google Drive, OneDrive,
// Dropbox, etc.). Implementations handle provider-specific APIs behind this
// uniform contract so the rest of the application treats all providers identically.
//
// Token refresh is handled internally by each implementation. Methods that use
// credentials return updated credentials alongside their result when a refresh
// occurs. The caller (connector service) is responsible for persisting updated
// credentials — the provider does not have database access.
package provider

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Provider is the contract all storage providers satisfy.
type Provider interface {
	// Type returns the provider type identifier (e.g., "google_drive", "onedrive").
	Type() string

	// AuthURL returns the OAuth authorization URL for connecting this provider.
	// The redirectURI is the frontend callback URL. The state parameter is a CSRF
	// token that the frontend validates on callback.
	AuthURL(ctx context.Context, redirectURI, state string) (string, error)

	// Exchange trades an authorization code for credentials.
	// The code comes from the OAuth callback. The redirectURI must match the one
	// used in AuthURL exactly (OAuth2 requirement).
	Exchange(ctx context.Context, code, redirectURI string) (*Credentials, error)

	// List returns entries (files and folders) at the given path.
	// Returns updated credentials if a token refresh occurred, nil otherwise.
	List(ctx context.Context, creds *Credentials, path string) ([]Entry, *Credentials, error)

	// Changes returns entries that changed since the given sync token.
	// The sync token is provider-specific and opaque to the caller.
	// Pass an empty token for the initial sync.
	// Returns updated credentials if a token refresh occurred, nil otherwise.
	Changes(ctx context.Context, creds *Credentials, path, syncToken string) ([]Change, string, *Credentials, error)

	// Fetch downloads a file's content by provider-specific reference.
	// The caller must close the returned ReadCloser.
	// Returns updated credentials if a token refresh occurred, nil otherwise.
	Fetch(ctx context.Context, creds *Credentials, ref string) (io.ReadCloser, *EntryMeta, *Credentials, error)
}

// WebhookProvider is an optional interface for providers that support
// push-based change notifications. Implementations that only support
// polling do not need to implement this.
type WebhookProvider interface {
	// RegisterWebhook sets up push notifications for changes at the given path.
	// The callbackURL is the connector's webhook endpoint.
	// Returns a registration that the connector uses to manage the subscription lifecycle.
	// Returns updated credentials if a token refresh occurred, nil otherwise.
	RegisterWebhook(ctx context.Context, creds *Credentials, callbackURL, path string) (*WebhookRegistration, *Credentials, error)

	// RenewWebhook extends or re-creates an existing webhook subscription.
	// Some providers (Microsoft Graph) require periodic renewal.
	// Returns updated credentials if a token refresh occurred, nil otherwise.
	RenewWebhook(ctx context.Context, creds *Credentials, registration *WebhookRegistration) (*WebhookRegistration, *Credentials, error)

	// VerifyWebhook validates an incoming webhook request from the provider.
	// Returns the provider-specific subscription ID for routing.
	// This is a static verification (signature check, challenge response) —
	// it does not require credentials.
	VerifyWebhook(r *http.Request) (subscriptionID string, err error)
}

// Credentials holds OAuth2 tokens for a provider connection.
// Serialized as JSON in the Provider model's encrypted Credentials field.
type Credentials struct {
	Extra        map[string]string `json:"extra,omitempty"`
	Expiry       time.Time         `json:"expiry"`
	AccessToken  string            `json:"access_token"`
	RefreshToken string            `json:"refresh_token"`
	TokenType    string            `json:"token_type"`
}

// Expired reports whether the access token has expired (with a 30s buffer).
func (c *Credentials) Expired() bool {
	if c.Expiry.IsZero() {
		return false
	}
	return time.Now().After(c.Expiry.Add(-30 * time.Second))
}

// Entry represents a file or folder in a provider's storage.
type Entry struct {
	// ModifiedAt is the last modification time.
	ModifiedAt time.Time `json:"modified_at"`

	// Ref is the provider-specific reference (file ID, path, etc.).
	// Used to fetch content or track changes.
	Ref string `json:"ref"`

	// Name is the display name of the file or folder.
	Name string `json:"name"`

	// MimeType is the MIME type of the file. Empty for folders.
	MimeType string `json:"mime_type,omitempty"`

	// ContentHash is a provider-specific content hash for deduplication.
	// Not all providers supply this (Google Docs don't have one).
	ContentHash string `json:"content_hash,omitempty"`

	// Size is the file size in bytes. Zero for folders.
	Size int64 `json:"size,omitempty"`

	// IsFolder indicates whether this entry is a folder.
	IsFolder bool `json:"is_folder"`
}

// EntryMeta is the metadata returned alongside fetched content.
type EntryMeta struct {
	Name        string `json:"name"`
	MimeType    string `json:"mime_type"`
	ContentHash string `json:"content_hash,omitempty"`
	Size        int64  `json:"size"`
}

// Change represents a detected change in a provider's storage.
type Change struct {
	// Entry is the file or folder that changed.
	// Nil if the entry was deleted.
	Entry *Entry `json:"entry,omitempty"`

	// Ref is the provider-specific reference for the changed item.
	// Always populated, even for deletions.
	Ref string `json:"ref"`

	// Type indicates what happened.
	Type ChangeType `json:"change_type"`
}

// ChangeType describes the kind of change detected.
type ChangeType string

// Change type values.
const (
	ChangeCreated  ChangeType = "created"
	ChangeModified ChangeType = "modified"
	ChangeDeleted  ChangeType = "deleted"
)

// WebhookRegistration holds the state for an active webhook subscription.
type WebhookRegistration struct {
	// Extra holds provider-specific registration metadata.
	Extra map[string]string `json:"extra,omitempty"`

	// Expiry is when the subscription expires and needs renewal.
	// Zero means no expiry.
	Expiry time.Time `json:"expiry,omitempty"`

	// ID is the provider-specific subscription ID.
	ID string `json:"id"`
}
