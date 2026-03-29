// Package onedrive implements the provider interface for OneDrive and SharePoint.
//
// Uses the Microsoft Graph API for file listing, change detection, and content
// retrieval. OAuth2 tokens are refreshed transparently — callers receive updated
// credentials when a refresh occurs.
package onedrive

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"time"

	"golang.org/x/oauth2"

	"github.com/zoobz-io/argus/provider"
)

const providerType = "onedrive"

// graphBaseURL is the default Microsoft Graph API base URL.
const graphBaseURL = "https://graph.microsoft.com/v1.0"

// microsoftAuthBase is the Microsoft identity platform base URL.
const microsoftAuthBase = "https://login.microsoftonline.com/common/oauth2/v2.0"

// validItemID matches Microsoft Graph item IDs (alphanumeric, hyphens, bangs).
var validItemID = regexp.MustCompile(`^[a-zA-Z0-9!_-]+$`)

// Config holds Microsoft OAuth2 application credentials.
// These are per-application, not per-tenant.
type Config struct {
	ClientID     string
	ClientSecret string
	Scopes       []string
}

// DefaultScopes returns the default OAuth2 scopes for OneDrive read access.
func DefaultScopes() []string {
	return []string{
		"Files.Read.All",
		"offline_access",
	}
}

// OneDrive implements provider.Provider for OneDrive and SharePoint.
type OneDrive struct {
	oauth    *oauth2.Config
	endpoint string // override for testing; empty uses default Graph API
}

// New creates a OneDrive provider with the given application credentials.
func New(cfg Config) *OneDrive {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}
	return &OneDrive{
		oauth: &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			Scopes:       scopes,
			Endpoint: oauth2.Endpoint{
				AuthURL:  microsoftAuthBase + "/authorize",
				TokenURL: microsoftAuthBase + "/token",
			},
		},
	}
}

// Type returns the provider type identifier.
func (o *OneDrive) Type() string {
	return providerType
}

// AuthURL returns the Microsoft OAuth2 authorization URL.
func (o *OneDrive) AuthURL(_ context.Context, redirectURI, state string) (string, error) {
	cfg := *o.oauth
	cfg.RedirectURL = redirectURI
	return cfg.AuthCodeURL(state, oauth2.AccessTypeOffline), nil
}

// Exchange trades an authorization code for credentials.
func (o *OneDrive) Exchange(ctx context.Context, code, redirectURI string) (*provider.Credentials, error) {
	cfg := *o.oauth
	cfg.RedirectURL = redirectURI
	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchanging auth code: %w", err)
	}
	return tokenToCreds(token), nil
}

// List returns files and folders at the given path (item ID).
// Pass "root" or an empty string for the root folder.
func (o *OneDrive) List(ctx context.Context, creds *provider.Credentials, path string) ([]provider.Entry, *provider.Credentials, error) {
	client, token, err := o.httpClient(ctx, creds)
	if err != nil {
		return nil, nil, err
	}

	itemID := path
	if itemID == "" {
		itemID = "root"
	}

	if itemID != "root" && !validItemID.MatchString(itemID) {
		return nil, nil, fmt.Errorf("invalid item ID: %s", itemID)
	}

	var entries []provider.Entry
	endpoint := o.graphURL("/me/drive/items/" + itemID + "/children")

	for endpoint != "" {
		var result listResponse
		if err := o.doJSON(ctx, client, endpoint, &result); err != nil {
			return nil, updatedCreds(creds, token), fmt.Errorf("listing files: %w", err)
		}

		for i := range result.Value {
			entries = append(entries, itemToEntry(&result.Value[i]))
		}

		endpoint = result.NextLink
	}

	return entries, updatedCreds(creds, token), nil
}

// Changes returns files that changed since the given sync token.
// Pass an empty token for the initial sync.
// The path parameter scopes the delta query to a specific folder path
// (e.g., "/Documents"). Pass empty for root-level delta.
func (o *OneDrive) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	client, token, err := o.httpClient(ctx, creds)
	if err != nil {
		return nil, syncToken, nil, err
	}

	// Build the delta URL. If we have a syncToken (deltaLink from previous call),
	// use it directly. Otherwise, build the initial delta URL.
	var endpoint string
	switch {
	case syncToken != "":
		endpoint = syncToken
	case path != "" && path != "/":
		endpoint = o.graphURL("/me/drive/root:/" + url.PathEscape(path) + ":/delta")
	default:
		endpoint = o.graphURL("/me/drive/root/delta")
	}

	var changes []provider.Change

	for {
		var result deltaResponse
		if err := o.doJSON(ctx, client, endpoint, &result); err != nil {
			return nil, syncToken, updatedCreds(creds, token), fmt.Errorf("querying delta: %w", err)
		}

		for i := range result.Value {
			change := itemToChange(&result.Value[i])
			changes = append(changes, change)
		}

		// If there is a deltaLink, we have consumed all pages.
		if result.DeltaLink != "" {
			return changes, result.DeltaLink, updatedCreds(creds, token), nil
		}

		// More pages to fetch.
		if result.NextLink != "" {
			endpoint = result.NextLink
			continue
		}

		// No deltaLink and no nextLink — shouldn't happen, but bail out safely.
		return changes, syncToken, updatedCreds(creds, token), nil
	}
}

// Fetch downloads a file's content by item ID.
func (o *OneDrive) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	client, token, err := o.httpClient(ctx, creds)
	if err != nil {
		return nil, nil, nil, err
	}

	if !validItemID.MatchString(ref) {
		return nil, nil, nil, fmt.Errorf("invalid item ID: %s", ref)
	}

	// Get file metadata first.
	var item graphItem
	metaURL := o.graphURL("/me/drive/items/" + ref)
	if err := o.doJSON(ctx, client, metaURL, &item); err != nil {
		return nil, nil, updatedCreds(creds, token), fmt.Errorf("getting file metadata: %w", err)
	}

	meta := &provider.EntryMeta{
		Name:        item.Name,
		MimeType:    mimeFromItem(&item),
		ContentHash: hashFromItem(&item),
		Size:        item.Size,
	}

	// Download content.
	contentURL := o.graphURL("/me/drive/items/" + ref + "/content")
	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, contentURL, nil)
	if reqErr != nil {
		return nil, nil, updatedCreds(creds, token), fmt.Errorf("creating content request: %w", reqErr)
	}

	resp, doErr := client.Do(req)
	if doErr != nil {
		return nil, nil, updatedCreds(creds, token), fmt.Errorf("downloading %s: %w", item.Name, doErr)
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, nil, updatedCreds(creds, token), fmt.Errorf("downloading %s: status %d", item.Name, resp.StatusCode)
	}

	return resp.Body, meta, updatedCreds(creds, token), nil
}

// --- Microsoft Graph API response types ---

// graphItem represents a OneDrive item from the Graph API.
type graphItem struct {
	LastModifiedAt time.Time    `json:"lastModifiedDateTime"`
	File           *graphFile   `json:"file,omitempty"`
	Folder         *graphFolder `json:"folder,omitempty"`
	Deleted        *struct{}    `json:"deleted,omitempty"`
	ID             string       `json:"id"`
	Name           string       `json:"name"`
	Size           int64        `json:"size"`
}

type graphFile struct {
	Hashes   *fileHashes `json:"hashes,omitempty"`
	MIMEType string      `json:"mimeType"`
}

type fileHashes struct {
	SHA1Hash   string `json:"sha1Hash,omitempty"`
	SHA256Hash string `json:"sha256Hash,omitempty"`
	QuickXOR   string `json:"quickXorHash,omitempty"`
}

type graphFolder struct {
	ChildCount int `json:"childCount"`
}

type listResponse struct {
	NextLink string      `json:"@odata.nextLink,omitempty"`
	Value    []graphItem `json:"value"`
}

type deltaResponse struct {
	NextLink  string      `json:"@odata.nextLink,omitempty"`
	DeltaLink string      `json:"@odata.deltaLink,omitempty"`
	Value     []graphItem `json:"value"`
}

// --- internal helpers ---

// httpClient creates an HTTP client with automatic token refresh.
func (o *OneDrive) httpClient(ctx context.Context, creds *provider.Credentials) (*http.Client, *oauth2.Token, error) {
	token := credsToToken(creds)
	tokenSource := o.oauth.TokenSource(ctx, token)

	// Get the potentially refreshed token.
	refreshed, err := tokenSource.Token()
	if err != nil {
		return nil, nil, fmt.Errorf("refreshing token: %w", err)
	}

	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(refreshed))
	return client, refreshed, nil
}

// graphURL returns the full URL for a Graph API path.
func (o *OneDrive) graphURL(path string) string {
	base := graphBaseURL
	if o.endpoint != "" {
		base = o.endpoint
	}
	return base + path
}

// doJSON performs a GET request and decodes the JSON response.
func (o *OneDrive) doJSON(ctx context.Context, client *http.Client, rawURL string, target any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(target); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}
	return nil
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

// itemToEntry converts a Graph API item to a provider.Entry.
func itemToEntry(item *graphItem) provider.Entry {
	return provider.Entry{
		Ref:         item.ID,
		Name:        item.Name,
		MimeType:    mimeFromItem(item),
		Size:        item.Size,
		ContentHash: hashFromItem(item),
		ModifiedAt:  item.LastModifiedAt,
		IsFolder:    item.Folder != nil,
	}
}

// itemToChange converts a Graph API delta item to a provider.Change.
func itemToChange(item *graphItem) provider.Change {
	if item.Deleted != nil {
		return provider.Change{
			Ref:  item.ID,
			Type: provider.ChangeDeleted,
		}
	}

	entry := itemToEntry(item)
	return provider.Change{
		Ref:   item.ID,
		Entry: &entry,
		Type:  provider.ChangeModified,
	}
}

// mimeFromItem extracts the MIME type from a Graph API item.
func mimeFromItem(item *graphItem) string {
	if item.File != nil {
		return item.File.MIMEType
	}
	if item.Folder != nil {
		return "application/vnd.ms-folder"
	}
	return ""
}

// hashFromItem extracts the best available hash from a Graph API item.
// Prefers SHA256 over SHA1 over QuickXOR.
func hashFromItem(item *graphItem) string {
	if item.File == nil || item.File.Hashes == nil {
		return ""
	}
	h := item.File.Hashes
	if h.SHA256Hash != "" {
		return h.SHA256Hash
	}
	if h.SHA1Hash != "" {
		return h.SHA1Hash
	}
	return h.QuickXOR
}
