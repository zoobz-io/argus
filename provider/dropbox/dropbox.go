// Package dropbox implements the provider interface for Dropbox.
//
// Uses the Dropbox API v2 for file listing, change detection, and content
// retrieval. OAuth2 tokens are refreshed transparently — callers receive updated
// credentials when a refresh occurs.
//
// All API calls use standard net/http — no Dropbox SDK dependency.
package dropbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/oauth2"

	"github.com/zoobz-io/argus/provider"
)

const providerType = "dropbox"

// API base URLs (overridden in tests via endpoint fields).
const (
	defaultAPIEndpoint     = "https://api.dropboxapi.com"
	defaultContentEndpoint = "https://content.dropboxapi.com"
	defaultAuthEndpoint    = "https://www.dropbox.com"
)

// Config holds Dropbox OAuth2 application credentials.
// These are per-application, not per-tenant.
type Config struct {
	ClientID     string
	ClientSecret string
}

// Dropbox implements provider.Provider for Dropbox.
type Dropbox struct {
	oauth           *oauth2.Config
	apiEndpoint     string // override for testing; empty uses default.
	contentEndpoint string // override for testing; empty uses default.
}

// New creates a Dropbox provider with the given application credentials.
func New(cfg Config) *Dropbox {
	return &Dropbox{
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Endpoint: oauth2.Endpoint{
				AuthURL:  defaultAuthEndpoint + "/oauth2/authorize",
				TokenURL: defaultAPIEndpoint + "/oauth2/token",
			},
		},
	}
}

// Type returns the provider type identifier.
func (d *Dropbox) Type() string {
	return providerType
}

// AuthURL returns the Dropbox OAuth2 authorization URL.
// Dropbox uses app-level permissions configured in the app console,
// so no scopes are included in the URL.
func (d *Dropbox) AuthURL(_ context.Context, redirectURI, state string) (string, error) {
	cfg := *d.oauth
	cfg.RedirectURL = redirectURI
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// Exchange trades an authorization code for credentials.
func (d *Dropbox) Exchange(ctx context.Context, code, redirectURI string) (*provider.Credentials, error) {
	cfg := *d.oauth
	cfg.RedirectURL = redirectURI
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	return tokenToCreds(token), nil
}

// List returns entries (files and folders) at the given path.
// Pass "" or "/" for the root folder.
func (d *Dropbox) List(ctx context.Context, creds *provider.Credentials, path string) ([]provider.Entry, *provider.Credentials, error) {
	client, refreshed, err := d.httpClient(ctx, creds)
	if err != nil {
		return nil, nil, err
	}

	folderPath := normalizePath(path)

	var entries []provider.Entry
	body := listFolderRequest{Path: folderPath}

	endpoint := d.apiURL() + "/2/files/list_folder"
	for {
		respBody, postErr := d.doJSON(ctx, client, endpoint, body)
		if postErr != nil {
			return nil, updatedCreds(creds, refreshed), fmt.Errorf("listing files: %w", postErr)
		}

		var result listFolderResponse
		if decErr := json.Unmarshal(respBody, &result); decErr != nil {
			return nil, updatedCreds(creds, refreshed), fmt.Errorf("decoding list response: %w", decErr)
		}

		for i := range result.Entries {
			entries = append(entries, metadataToEntry(&result.Entries[i]))
		}

		if !result.HasMore {
			break
		}

		// Continue with cursor.
		body = listFolderRequest{Cursor: result.Cursor}
		endpoint = d.apiURL() + "/2/files/list_folder/continue"
	}

	return entries, updatedCreds(creds, refreshed), nil
}

// Changes returns entries that changed since the given sync token (cursor).
// Pass an empty token for the initial sync — returns a cursor with no changes.
func (d *Dropbox) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	client, refreshed, err := d.httpClient(ctx, creds)
	if err != nil {
		return nil, syncToken, nil, err
	}

	// Initial sync: get a cursor by listing the folder.
	if syncToken == "" {
		cursor, listErr := d.getLatestCursor(ctx, client, path)
		if listErr != nil {
			return nil, "", updatedCreds(creds, refreshed), fmt.Errorf("getting initial cursor: %w", listErr)
		}
		return nil, cursor, updatedCreds(creds, refreshed), nil
	}

	// Continue from cursor to get changes.
	var changes []provider.Change
	cursor := syncToken

	for {
		reqBody := listFolderContinueRequest{Cursor: cursor}
		respBody, postErr := d.doJSON(ctx, client, d.apiURL()+"/2/files/list_folder/continue", reqBody)
		if postErr != nil {
			return nil, syncToken, updatedCreds(creds, refreshed), fmt.Errorf("listing changes: %w", postErr)
		}

		var result listFolderResponse
		if decErr := json.Unmarshal(respBody, &result); decErr != nil {
			return nil, syncToken, updatedCreds(creds, refreshed), fmt.Errorf("decoding changes response: %w", decErr)
		}

		for i := range result.Entries {
			change := metadataToChange(&result.Entries[i])
			if change != nil {
				changes = append(changes, *change)
			}
		}

		cursor = result.Cursor
		if !result.HasMore {
			break
		}
	}

	return changes, cursor, updatedCreds(creds, refreshed), nil
}

// Fetch downloads a file's content by path.
func (d *Dropbox) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	client, refreshed, err := d.httpClient(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	if ref == "" {
		return nil, nil, updatedCreds(creds, refreshed), fmt.Errorf("empty file reference")
	}

	// Dropbox download uses Dropbox-API-Arg header for parameters.
	apiArg, marshalErr := json.Marshal(map[string]string{"path": ref})
	if marshalErr != nil {
		return nil, nil, updatedCreds(creds, refreshed), fmt.Errorf("marshaling api arg: %w", marshalErr)
	}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, d.contentURL()+"/2/files/download", nil)
	if reqErr != nil {
		return nil, nil, updatedCreds(creds, refreshed), fmt.Errorf("creating download request: %w", reqErr)
	}
	req.Header.Set("Dropbox-API-Arg", string(apiArg))

	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, nil, updatedCreds(creds, refreshed), fmt.Errorf("downloading file: %w", doErr)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, nil, updatedCreds(creds, refreshed), fmt.Errorf("download returned status %d", resp.StatusCode)
	}

	// Parse metadata from the Dropbox-API-Result header.
	var meta fileMetadata
	resultHeader := resp.Header.Get("Dropbox-API-Result")
	if resultHeader != "" {
		if decErr := json.Unmarshal([]byte(resultHeader), &meta); decErr != nil {
			_ = resp.Body.Close()
			return nil, nil, updatedCreds(creds, refreshed), fmt.Errorf("decoding download metadata: %w", decErr)
		}
	}

	entryMeta := &provider.EntryMeta{
		Name:        meta.Name,
		MimeType:    guessMIME(meta.Name),
		ContentHash: meta.ContentHash,
		Size:        meta.Size,
	}

	return resp.Body, entryMeta, updatedCreds(creds, refreshed), nil
}

// --- internal helpers ---

// httpClient creates an HTTP client with automatic token refresh.
func (d *Dropbox) httpClient(ctx context.Context, creds *provider.Credentials) (*http.Client, *oauth2.Token, error) {
	token := credsToToken(creds)
	tokenSource := d.oauth.TokenSource(ctx, token)

	// Get the potentially refreshed token.
	refreshed, err := tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("refreshing token: %w", err)
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(refreshed))
	return client, refreshed, nil
}

// doJSON sends a POST request with JSON body and returns the response bytes.
func (d *Dropbox) doJSON(ctx context.Context, client *http.Client, url string, body any) ([]byte, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// getLatestCursor calls list_folder/get_latest_cursor to get a cursor
// without listing all entries.
func (d *Dropbox) getLatestCursor(ctx context.Context, client *http.Client, path string) (string, error) {
	folderPath := normalizePath(path)
	body := listFolderRequest{Path: folderPath}

	respBody, err := d.doJSON(ctx, client, d.apiURL()+"/2/files/list_folder/get_latest_cursor", body)
	if err != nil {
		return "", err
	}

	var result struct {
		Cursor string `json:"cursor"`
	}
	if decErr := json.Unmarshal(respBody, &result); decErr != nil {
		return "", fmt.Errorf("decoding cursor response: %w", decErr)
	}
	return result.Cursor, nil
}

func (d *Dropbox) apiURL() string {
	if d.apiEndpoint != "" {
		return d.apiEndpoint
	}
	return defaultAPIEndpoint
}

func (d *Dropbox) contentURL() string {
	if d.contentEndpoint != "" {
		return d.contentEndpoint
	}
	return defaultContentEndpoint
}

// normalizePath converts paths for the Dropbox API.
// Empty string and "/" map to "" (root). All other paths must start with "/".
func normalizePath(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

// --- token conversion helpers ---

func tokenToCreds(t *oauth2.Token) *provider.Credentials {
	return &provider.Credentials{
		AccessToken:  t.AccessToken,
		RefreshToken: t.RefreshToken,
		TokenType:    t.TokenType,
		Expiry:       t.Expiry,
	}
}

func credsToToken(c *provider.Credentials) *oauth2.Token {
	return &oauth2.Token{
		AccessToken:  c.AccessToken,
		RefreshToken: c.RefreshToken,
		TokenType:    c.TokenType,
		Expiry:       c.Expiry,
	}
}

func updatedCreds(original *provider.Credentials, current *oauth2.Token) *provider.Credentials {
	if current == nil || current.AccessToken == original.AccessToken {
		return nil
	}
	return tokenToCreds(current)
}

// --- Dropbox API types ---

type listFolderRequest struct {
	Path   string `json:"path,omitempty"`
	Cursor string `json:"cursor,omitempty"`
}

type listFolderContinueRequest struct {
	Cursor string `json:"cursor"`
}

type listFolderResponse struct {
	Cursor  string         `json:"cursor"`
	Entries []fileMetadata `json:"entries"`
	HasMore bool           `json:"has_more"`
}

type fileMetadata struct {
	Tag            string `json:".tag"`
	Name           string `json:"name"`
	PathLower      string `json:"path_lower"`
	PathDisplay    string `json:"path_display"`
	ID             string `json:"id"`
	ContentHash    string `json:"content_hash,omitempty"`
	ServerModified string `json:"server_modified,omitempty"`
	Size           int64  `json:"size,omitempty"`
}

// --- entry conversion ---

func metadataToEntry(m *fileMetadata) provider.Entry {
	isFolder := m.Tag == "folder"
	var modTime time.Time
	if m.ServerModified != "" {
		modTime, _ = time.Parse(time.RFC3339, m.ServerModified)
	}
	mimeType := ""
	if !isFolder {
		mimeType = guessMIME(m.Name)
	}
	return provider.Entry{
		Ref:         m.PathDisplay,
		Name:        m.Name,
		MimeType:    mimeType,
		Size:        m.Size,
		ContentHash: m.ContentHash,
		ModifiedAt:  modTime,
		IsFolder:    isFolder,
	}
}

func metadataToChange(m *fileMetadata) *provider.Change {
	if m.Tag == "deleted" {
		return &provider.Change{
			Ref:  m.PathDisplay,
			Type: provider.ChangeDeleted,
		}
	}

	entry := metadataToEntry(m)
	return &provider.Change{
		Ref:   m.PathDisplay,
		Entry: &entry,
		Type:  provider.ChangeModified,
	}
}

// guessMIME returns a MIME type based on file extension.
// Dropbox does not return MIME types in metadata, so we infer from the name.
func guessMIME(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.HasSuffix(lower, ".pdf"):
		return "application/pdf"
	case strings.HasSuffix(lower, ".doc"):
		return "application/msword"
	case strings.HasSuffix(lower, ".docx"):
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case strings.HasSuffix(lower, ".xls"):
		return "application/vnd.ms-excel"
	case strings.HasSuffix(lower, ".xlsx"):
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case strings.HasSuffix(lower, ".pptx"):
		return "application/vnd.openxmlformats-officedocument.presentationml.presentation"
	case strings.HasSuffix(lower, ".txt"):
		return "text/plain"
	case strings.HasSuffix(lower, ".csv"):
		return "text/csv"
	case strings.HasSuffix(lower, ".json"):
		return "application/json"
	case strings.HasSuffix(lower, ".png"):
		return "image/png"
	case strings.HasSuffix(lower, ".jpg"), strings.HasSuffix(lower, ".jpeg"):
		return "image/jpeg"
	default:
		return "application/octet-stream"
	}
}
