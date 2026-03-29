// Package gcs implements the provider interface for Google Cloud Storage.
//
// Uses the Google Cloud Storage JSON API for object listing, change detection,
// and content retrieval. Authentication is via service account credentials (JSON
// key file) — there is no OAuth2 user consent flow. The service account handles
// its own token lifecycle internally.
package gcs

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/auth/credentials"
	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/zoobz-io/argus/provider"
)

const providerType = "gcs"

// GCS implements provider.Provider for Google Cloud Storage.
type GCS struct {
	endpoint string // override for testing; empty uses default GCS API
}

// New creates a GCS provider.
func New() *GCS {
	return &GCS{}
}

// Type returns the provider type identifier.
func (g *GCS) Type() string {
	return providerType
}

// AuthURL returns an error — GCS uses service account credentials, not OAuth.
func (g *GCS) AuthURL(_ context.Context, _, _ string) (string, error) {
	return "", fmt.Errorf("GCS uses service account credentials, not OAuth")
}

// Exchange returns an error — GCS uses service account credentials, not OAuth.
func (g *GCS) Exchange(_ context.Context, _, _ string) (*provider.Credentials, error) {
	return nil, fmt.Errorf("GCS uses service account credentials, not OAuth")
}

// List returns objects and prefixes (pseudo-folders) at the given path prefix.
// Returns nil credentials since service accounts handle their own token lifecycle.
func (g *GCS) List(ctx context.Context, creds *provider.Credentials, path string) ([]provider.Entry, *provider.Credentials, error) {
	bucket, prefix, client, err := g.setup(ctx, creds, path)
	if err != nil {
		return nil, nil, err
	}
	defer func() { _ = client.Close() }()

	query := &storage.Query{
		Prefix:    prefix,
		Delimiter: "/",
	}

	var entries []provider.Entry
	it := client.Bucket(bucket).Objects(ctx, query)

	for {
		attrs, iterErr := it.Next()
		if errors.Is(iterErr, iterator.Done) {
			break
		}
		if iterErr != nil {
			return nil, nil, fmt.Errorf("listing objects: %w", iterErr)
		}

		entries = append(entries, attrsToEntry(attrs))
	}

	return entries, nil, nil
}

// Changes returns objects that changed since the given sync token.
// The sync token is an RFC3339 timestamp representing the last sync time.
// Pass an empty token for the initial sync (returns current timestamp without changes).
// Returns nil credentials since service accounts handle their own token lifecycle.
func (g *GCS) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	now := time.Now().UTC()

	// Initial sync — return current timestamp as the starting token.
	if syncToken == "" {
		return nil, now.Format(time.RFC3339), nil, nil
	}

	updatedAfter, parseErr := time.Parse(time.RFC3339, syncToken)
	if parseErr != nil {
		return nil, syncToken, nil, fmt.Errorf("parsing sync token: %w", parseErr)
	}

	bucket, prefix, client, err := g.setup(ctx, creds, path)
	if err != nil {
		return nil, syncToken, nil, err
	}
	defer func() { _ = client.Close() }()

	query := &storage.Query{
		Prefix: prefix,
	}

	var changes []provider.Change
	it := client.Bucket(bucket).Objects(ctx, query)

	for {
		attrs, iterErr := it.Next()
		if errors.Is(iterErr, iterator.Done) {
			break
		}
		if iterErr != nil {
			return nil, syncToken, nil, fmt.Errorf("listing objects for changes: %w", iterErr)
		}

		// Only include objects updated after the sync token timestamp.
		if attrs.Prefix != "" {
			continue
		}
		if !attrs.Updated.After(updatedAfter) {
			continue
		}

		entry := attrsToEntry(attrs)
		changeType := provider.ChangeModified
		if attrs.Created.After(updatedAfter) {
			changeType = provider.ChangeCreated
		}

		changes = append(changes, provider.Change{
			Ref:   attrs.Name,
			Entry: &entry,
			Type:  changeType,
		})
	}

	return changes, now.Format(time.RFC3339), nil, nil
}

// Fetch downloads an object's content by object key.
// Returns nil credentials since service accounts handle their own token lifecycle.
func (g *GCS) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	bucketName, _, client, err := g.setup(ctx, creds, "")
	if err != nil {
		return nil, nil, nil, err
	}

	obj := client.Bucket(bucketName).Object(ref)

	attrs, attrErr := obj.Attrs(ctx)
	if attrErr != nil {
		_ = client.Close()
		return nil, nil, nil, fmt.Errorf("getting object attributes: %w", attrErr)
	}

	meta := &provider.EntryMeta{
		Name:        objectName(ref),
		MimeType:    attrs.ContentType,
		ContentHash: fmt.Sprintf("%x", attrs.MD5),
		Size:        attrs.Size,
	}

	reader, readErr := obj.NewReader(ctx)
	if readErr != nil {
		_ = client.Close()
		return nil, nil, nil, fmt.Errorf("reading object: %w", readErr)
	}

	// Wrap the reader so the client is closed when the reader is closed.
	return &clientClosingReader{ReadCloser: reader, client: client}, meta, nil, nil
}

// setup validates credentials and creates a storage client.
func (g *GCS) setup(ctx context.Context, creds *provider.Credentials, path string) (bucket, prefix string, client *storage.Client, err error) {
	if creds == nil {
		return "", "", nil, fmt.Errorf("credentials required")
	}
	if creds.Extra == nil {
		return "", "", nil, fmt.Errorf("credentials missing extra fields")
	}

	bucket = creds.Extra["bucket"]
	if bucket == "" {
		return "", "", nil, fmt.Errorf("bucket not specified in credentials")
	}

	credsJSON, decodeErr := decodeCredentialsJSON(creds.Extra["credentials_json"])
	if decodeErr != nil {
		return "", "", nil, decodeErr
	}

	var opts []option.ClientOption
	if g.endpoint != "" {
		// Test mode: use endpoint override without real credentials.
		opts = append(opts, option.WithEndpoint(g.endpoint), option.WithoutAuthentication())
	} else {
		googleCreds, credErr := credentials.DetectDefault(&credentials.DetectOptions{
			Scopes:          []string{storage.ScopeReadOnly},
			CredentialsJSON: credsJSON,
		})
		if credErr != nil {
			return "", "", nil, fmt.Errorf("parsing service account credentials: %w", credErr)
		}
		opts = append(opts, option.WithAuthCredentials(googleCreds))
	}

	client, clientErr := storage.NewClient(ctx, opts...)
	if clientErr != nil {
		return "", "", nil, fmt.Errorf("creating storage client: %w", clientErr)
	}

	prefix = path
	return bucket, prefix, client, nil
}

// decodeCredentialsJSON decodes the base64-encoded service account JSON key.
func decodeCredentialsJSON(encoded string) ([]byte, error) {
	if encoded == "" {
		return nil, fmt.Errorf("credentials_json not specified in credentials")
	}
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("decoding credentials_json: %w", err)
	}
	return decoded, nil
}

// attrsToEntry converts GCS object attributes to a provider.Entry.
func attrsToEntry(attrs *storage.ObjectAttrs) provider.Entry {
	// Prefix-only results represent pseudo-folders.
	if attrs.Prefix != "" {
		name := strings.TrimSuffix(attrs.Prefix, "/")
		if idx := strings.LastIndex(name, "/"); idx >= 0 {
			name = name[idx+1:]
		}
		return provider.Entry{
			Ref:      attrs.Prefix,
			Name:     name,
			IsFolder: true,
		}
	}

	return provider.Entry{
		Ref:         attrs.Name,
		Name:        objectName(attrs.Name),
		MimeType:    attrs.ContentType,
		Size:        attrs.Size,
		ContentHash: fmt.Sprintf("%x", attrs.MD5),
		ModifiedAt:  attrs.Updated,
		IsFolder:    false,
	}
}

// objectName extracts the filename from a full object key.
func objectName(key string) string {
	key = strings.TrimSuffix(key, "/")
	if idx := strings.LastIndex(key, "/"); idx >= 0 {
		return key[idx+1:]
	}
	return key
}

// syncTokenToUnix converts a sync token (RFC3339 timestamp) to a Unix timestamp string.
func syncTokenToUnix(token string) string {
	t, err := time.Parse(time.RFC3339, token)
	if err != nil {
		return "0"
	}
	return strconv.FormatInt(t.Unix(), 10)
}

// clientClosingReader wraps a ReadCloser and closes the storage client when done.
type clientClosingReader struct {
	io.ReadCloser
	client *storage.Client
}

// Close closes the underlying reader and the storage client.
func (r *clientClosingReader) Close() error {
	readErr := r.ReadCloser.Close()
	clientErr := r.client.Close()
	if readErr != nil {
		return readErr
	}
	return clientErr
}
