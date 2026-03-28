// Package googledrive implements the provider interface for Google Drive.
//
// Uses the Google Drive API v3 for file listing, change detection, and content
// retrieval. OAuth2 tokens are refreshed transparently — callers receive updated
// credentials when a refresh occurs.
package googledrive

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/zoobz-io/argus/provider"
)

const providerType = "google_drive"

// validFolderID matches Google Drive file/folder IDs.
var validFolderID = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// Config holds Google OAuth2 application credentials.
// These are per-application, not per-tenant.
type Config struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// DefaultScopes returns the default OAuth2 scopes for Google Drive read access.
func DefaultScopes() []string {
	return []string{
		drive.DriveReadonlyScope,
	}
}

// GoogleDrive implements provider.Provider for Google Drive.
type GoogleDrive struct {
	oauth    *oauth2.Config
	endpoint string // override for testing; empty uses default Google API
}

// New creates a Google Drive provider with the given application credentials.
func New(cfg Config) *GoogleDrive {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}
	return &GoogleDrive{
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       scopes,
			Endpoint:     google.Endpoint,
		},
	}
}

// Type returns the provider type identifier.
func (g *GoogleDrive) Type() string {
	return providerType
}

// AuthURL returns the Google OAuth2 authorization URL.
func (g *GoogleDrive) AuthURL(_ context.Context, redirectURI, state string) (string, error) {
	cfg := *g.oauth
	cfg.RedirectURL = redirectURI
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce), nil
}

// Exchange trades an authorization code for credentials.
func (g *GoogleDrive) Exchange(ctx context.Context, code, redirectURI string) (*provider.Credentials, error) {
	cfg := *g.oauth
	cfg.RedirectURL = redirectURI
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	return tokenToCreds(token), nil
}

// List returns files and folders at the given path (folder ID).
// Pass "root" or an empty string for the root folder.
func (g *GoogleDrive) List(ctx context.Context, creds *provider.Credentials, path string) ([]provider.Entry, *provider.Credentials, error) {
	svc, token, err := g.driveService(ctx, creds)
	if err != nil {
		return nil, nil, err
	}

	folderID := path
	if folderID == "" {
		folderID = "root"
	}
	if !validFolderID.MatchString(folderID) {
		return nil, nil, fmt.Errorf("invalid folder ID: %s", folderID)
	}

	query := fmt.Sprintf("'%s' in parents and trashed = false", folderID)
	var entries []provider.Entry
	var pageToken string

	for {
		call := svc.Files.List().
			Q(query).
			Fields("nextPageToken, files(id, name, mimeType, size, md5Checksum, modifiedTime)").
			PageSize(1000)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		result, listErr := call.Do()
		if listErr != nil {
			return nil, updatedCreds(creds, token), fmt.Errorf("listing files: %w", listErr)
		}

		for _, f := range result.Files {
			entries = append(entries, fileToEntry(f))
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	return entries, updatedCreds(creds, token), nil
}

// Changes returns files that changed since the given sync token.
// Pass an empty token for the initial sync (returns a start token without changes).
func (g *GoogleDrive) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	svc, token, err := g.driveService(ctx, creds)
	if err != nil {
		return nil, syncToken, nil, err
	}

	// Get initial start token if none provided.
	if syncToken == "" {
		startToken, tokenErr := svc.Changes.GetStartPageToken().Do()
		if tokenErr != nil {
			return nil, "", updatedCreds(creds, token), fmt.Errorf("getting start token: %w", tokenErr)
		}
		return nil, startToken.StartPageToken, updatedCreds(creds, token), nil
	}

	var changes []provider.Change
	pageToken := syncToken

	for {
		result, changesErr := svc.Changes.List(pageToken).
			Fields("nextPageToken, newStartPageToken, changes(fileId, removed, file(id, name, mimeType, size, md5Checksum, modifiedTime, parents))").
			PageSize(1000).
			IncludeRemoved(true).
			Do()
		if changesErr != nil {
			return nil, syncToken, updatedCreds(creds, token), fmt.Errorf("listing changes: %w", changesErr)
		}

		for _, c := range result.Changes {
			change := changeToProviderChange(c, path)
			if change != nil {
				changes = append(changes, *change)
			}
		}

		if result.NewStartPageToken != "" {
			return changes, result.NewStartPageToken, updatedCreds(creds, token), nil
		}
		pageToken = result.NextPageToken
	}
}

// Fetch downloads a file's content by file ID.
func (g *GoogleDrive) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	svc, token, err := g.driveService(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	// Get file metadata first.
	file, metaErr := svc.Files.Get(ref).
		Fields("id, name, mimeType, size, md5Checksum").
		Do()
	if metaErr != nil {
		return nil, nil, updatedCreds(creds, token), fmt.Errorf("getting file metadata: %w", metaErr)
	}

	meta := &provider.EntryMeta{
		Name:        file.Name,
		MimeType:    file.MimeType,
		ContentHash: file.Md5Checksum,
		Size:        file.Size,
	}

	// Google Docs/Sheets/Slides need export, not download.
	if isGoogleNative(file.MimeType) {
		exportMime := googleExportMIME(file.MimeType)
		resp, exportErr := svc.Files.Export(ref, exportMime).Download()
		if exportErr != nil {
			return nil, nil, updatedCreds(creds, token), fmt.Errorf("exporting %s as %s: %w", file.Name, exportMime, exportErr)
		}
		meta.MimeType = exportMime
		return resp.Body, meta, updatedCreds(creds, token), nil
	}

	resp, dlErr := svc.Files.Get(ref).Download()
	if dlErr != nil {
		return nil, nil, updatedCreds(creds, token), fmt.Errorf("downloading %s: %w", file.Name, dlErr)
	}
	return resp.Body, meta, updatedCreds(creds, token), nil
}

// driveService creates a Drive API service with automatic token refresh.
func (g *GoogleDrive) driveService(ctx context.Context, creds *provider.Credentials) (*drive.Service, *oauth2.Token, error) {
	token := credsToToken(creds)
	tokenSource := g.oauth.TokenSource(ctx, token)

	opts := []option.ClientOption{option.WithTokenSource(tokenSource)}
	if g.endpoint != "" {
		opts = append(opts, option.WithEndpoint(g.endpoint))
	}
	svc, err := drive.NewService(ctx, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("creating drive service: %w", err)
	}

	// Get the potentially refreshed token.
	refreshed, err := tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("refreshing token: %w", err)
	}

	return svc, refreshed, nil
}

// tokenToCreds converts an oauth2.Token to provider.Credentials.
func tokenToCreds(t *oauth2.Token) *provider.Credentials {
	return &provider.Credentials{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
	}
}

// credsToToken converts provider.Credentials to an oauth2.Token.
func credsToToken(c *provider.Credentials) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		TokenType:    c.TokenType,
		Expiry:       c.Expiry,
	}
}

// updatedCreds returns updated credentials if the token changed, nil otherwise.
func updatedCreds(original *provider.Credentials, current *oauth2.Token) *provider.Credentials {
	if current == nil || current.AccessToken == original.AccessToken {
		return nil
	}
	return tokenToCreds(current)
}

// fileToEntry converts a Google Drive File to a provider.Entry.
func fileToEntry(f *drive.File) provider.Entry {
	modTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)
	return provider.Entry{
		Ref:         f.Id,
		Name:        f.Name,
		MimeType:    f.MimeType,
		Size:        f.Size,
		ContentHash: f.Md5Checksum,
		ModifiedAt:  modTime,
		IsFolder:    f.MimeType == "application/vnd.google-apps.folder",
	}
}

// changeToProviderChange converts a Google Drive Change to a provider.Change.
// Returns nil if the change is not relevant to the given path (folder ID).
// Pass empty path to include all changes.
//
// NOTE: Parent filtering is shallow — only checks immediate parents.
// Files in subfolders of a watched path are not detected. Recursive
// watch would require maintaining a folder tree per watched path.
func changeToProviderChange(c *drive.Change, path string) *provider.Change {
	if c.Removed {
		return &provider.Change{
			Ref:  c.FileId,
			Type: provider.ChangeDeleted,
		}
	}

	if c.File == nil {
		return nil
	}

	// Filter by parent folder if path is specified.
	if path != "" && path != "root" {
		found := false
		for _, parent := range c.File.Parents {
			if parent == path {
				found = true
				break
			}
		}
		if !found {
			return nil
		}
	}

	entry := fileToEntry(c.File)
	return &provider.Change{
		Ref:   c.FileId,
		Entry: &entry,
		Type:  provider.ChangeModified,
	}
}

// isGoogleNative returns true for Google Docs, Sheets, Slides, etc.
func isGoogleNative(mimeType string) bool {
	return strings.HasPrefix(mimeType, "application/vnd.google-apps.")
}

// googleExportMIME returns the export MIME type for a Google native format.
func googleExportMIME(mimeType string) string {
	switch mimeType {
	case "application/vnd.google-apps.document":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case "application/vnd.google-apps.spreadsheet":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case "application/vnd.google-apps.presentation":
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	default:
		return "application/pdf"
	}
}
