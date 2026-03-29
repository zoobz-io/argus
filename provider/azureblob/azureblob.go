// Package azureblob implements the provider interface for Azure Blob Storage.
//
// Uses the Azure Blob Storage REST API with SharedKey authentication.
// No Azure SDK dependency — all operations are plain HTTP calls.
// Azure Blob uses static credentials (account name + account key), not OAuth2.
package azureblob

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/zoobz-io/argus/provider"
)

const providerType = "azure_blob"

// AzureBlob implements provider.Provider for Azure Blob Storage.
type AzureBlob struct {
	client *http.Client
}

// New creates an Azure Blob Storage provider.
func New() *AzureBlob {
	return &AzureBlob{
		client: http.DefaultClient,
	}
}

// Type returns the provider type identifier.
func (a *AzureBlob) Type() string {
	return providerType
}

// AuthURL returns an error because Azure Blob uses static credentials, not OAuth.
func (a *AzureBlob) AuthURL(_ context.Context, _, _ string) (string, error) {
	return "", fmt.Errorf("azure blob storage uses static credentials, not OAuth")
}

// Exchange returns an error because Azure Blob uses static credentials, not OAuth.
func (a *AzureBlob) Exchange(_ context.Context, _, _ string) (*provider.Credentials, error) {
	return nil, fmt.Errorf("azure blob storage uses static credentials, not OAuth")
}

// List returns blobs and virtual directories at the given prefix path.
// Pass an empty string for the root of the container.
func (a *AzureBlob) List(ctx context.Context, creds *provider.Credentials, path string) ([]provider.Entry, *provider.Credentials, error) {
	endpoint, container, err := a.resolveEndpoint(creds)
	if err != nil {
		return nil, nil, err
	}

	prefix := normalizePath(path)

	baseURL := fmt.Sprintf("%s/%s", endpoint, container)
	params := url.Values{}
	params.Set("restype", "container")
	params.Set("comp", "list")
	params.Set("delimiter", "/")
	if prefix != "" {
		params.Set("prefix", prefix)
	}

	var entries []provider.Entry
	marker := ""

	for {
		qp := make(url.Values)
		for k, v := range params {
			qp[k] = v
		}
		if marker != "" {
			qp.Set("marker", marker)
		}

		reqURL := baseURL + "?" + qp.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, nil, fmt.Errorf("creating list request: %w", err)
		}

		if signErr := a.signRequest(req, creds); signErr != nil {
			return nil, nil, fmt.Errorf("signing list request: %w", signErr)
		}

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, nil, fmt.Errorf("listing blobs: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, nil, fmt.Errorf("list blobs: status %d: %s", resp.StatusCode, string(body))
		}

		var result listBlobsResult
		if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return nil, nil, fmt.Errorf("decoding list response: %w", err)
		}
		_ = resp.Body.Close()

		// Virtual directories (prefixes).
		for _, bp := range result.Blobs.BlobPrefixes {
			name := strings.TrimSuffix(bp.Name, "/")
			if i := strings.LastIndex(name, "/"); i >= 0 {
				name = name[i+1:]
			}
			entries = append(entries, provider.Entry{
				Ref:      bp.Name,
				Name:     name,
				IsFolder: true,
			})
		}

		// Blobs.
		for _, b := range result.Blobs.Blobs {
			name := b.Name
			if i := strings.LastIndex(name, "/"); i >= 0 {
				name = name[i+1:]
			}
			var modTime time.Time
			if b.Properties.LastModified != "" {
				modTime, _ = time.Parse(time.RFC1123, b.Properties.LastModified)
			}
			entries = append(entries, provider.Entry{
				Ref:         b.Name,
				Name:        name,
				MimeType:    b.Properties.ContentType,
				Size:        b.Properties.ContentLength,
				ContentHash: b.Properties.ContentMD5,
				ModifiedAt:  modTime,
			})
		}

		if result.NextMarker == "" {
			break
		}
		marker = result.NextMarker
	}

	return entries, nil, nil
}

// Changes returns blobs that changed since the given sync token.
// The sync token is an RFC3339 timestamp of the last sync.
// Pass an empty token for the initial sync (returns current timestamp without changes).
func (a *AzureBlob) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	// Initial sync: return current timestamp as the starting token.
	if syncToken == "" {
		return nil, now, nil, nil
	}

	since, err := time.Parse(time.RFC3339, syncToken)
	if err != nil {
		return nil, syncToken, nil, fmt.Errorf("parsing sync token: %w", err)
	}

	endpoint, container, err := a.resolveEndpoint(creds)
	if err != nil {
		return nil, syncToken, nil, err
	}

	prefix := normalizePath(path)

	baseURL := fmt.Sprintf("%s/%s", endpoint, container)
	params := url.Values{}
	params.Set("restype", "container")
	params.Set("comp", "list")
	if prefix != "" {
		params.Set("prefix", prefix)
	}

	var changes []provider.Change
	marker := ""

	for {
		qp := make(url.Values)
		for k, v := range params {
			qp[k] = v
		}
		if marker != "" {
			qp.Set("marker", marker)
		}

		reqURL := baseURL + "?" + qp.Encode()
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, syncToken, nil, fmt.Errorf("creating changes request: %w", err)
		}

		if signErr := a.signRequest(req, creds); signErr != nil {
			return nil, syncToken, nil, fmt.Errorf("signing changes request: %w", signErr)
		}

		resp, err := a.client.Do(req)
		if err != nil {
			return nil, syncToken, nil, fmt.Errorf("listing changes: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			_ = resp.Body.Close()
			return nil, syncToken, nil, fmt.Errorf("list changes: status %d: %s", resp.StatusCode, string(body))
		}

		var result listBlobsResult
		if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
			_ = resp.Body.Close()
			return nil, syncToken, nil, fmt.Errorf("decoding changes response: %w", err)
		}
		_ = resp.Body.Close()

		for _, b := range result.Blobs.Blobs {
			var modTime time.Time
			if b.Properties.LastModified != "" {
				modTime, _ = time.Parse(time.RFC1123, b.Properties.LastModified)
			}

			if !modTime.After(since) {
				continue
			}

			name := b.Name
			if i := strings.LastIndex(name, "/"); i >= 0 {
				name = name[i+1:]
			}

			entry := provider.Entry{
				Ref:         b.Name,
				Name:        name,
				MimeType:    b.Properties.ContentType,
				Size:        b.Properties.ContentLength,
				ContentHash: b.Properties.ContentMD5,
				ModifiedAt:  modTime,
			}
			changes = append(changes, provider.Change{
				Ref:   b.Name,
				Entry: &entry,
				Type:  provider.ChangeModified,
			})
		}

		if result.NextMarker == "" {
			break
		}
		marker = result.NextMarker
	}

	return changes, now, nil, nil
}

// Fetch downloads a blob's content by its full blob name (path).
func (a *AzureBlob) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	endpoint, container, err := a.resolveEndpoint(creds)
	if err != nil {
		return nil, nil, nil, err
	}

	fetchURL := fmt.Sprintf("%s/%s/%s", endpoint, container, escapeRefPath(ref))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fetchURL, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating fetch request: %w", err)
	}

	if signErr := a.signRequest(req, creds); signErr != nil {
		return nil, nil, nil, fmt.Errorf("signing fetch request: %w", signErr)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetching blob: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, nil, nil, fmt.Errorf("fetch blob: status %d: %s", resp.StatusCode, string(body))
	}

	name := ref
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}

	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	meta := &provider.EntryMeta{
		Name:        name,
		MimeType:    resp.Header.Get("Content-Type"),
		ContentHash: resp.Header.Get("Content-MD5"),
		Size:        size,
	}

	return resp.Body, meta, nil, nil
}

// signRequest signs an HTTP request using Azure SharedKey authentication.
func (a *AzureBlob) signRequest(req *http.Request, creds *provider.Credentials) error {
	account := creds.AccessToken
	key := creds.RefreshToken

	keyBytes, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return fmt.Errorf("decoding account key: %w", err)
	}

	now := time.Now().UTC().Format(http.TimeFormat)
	req.Header.Set("x-ms-date", now)
	req.Header.Set("x-ms-version", "2020-10-02")

	contentLength := ""
	if req.ContentLength > 0 {
		contentLength = strconv.FormatInt(req.ContentLength, 10)
	}

	stringToSign := strings.Join([]string{
		req.Method,
		req.Header.Get("Content-Encoding"),
		req.Header.Get("Content-Language"),
		contentLength,
		req.Header.Get("Content-MD5"),
		req.Header.Get("Content-Type"),
		"", // Date (empty when using x-ms-date).
		req.Header.Get("If-Modified-Since"),
		req.Header.Get("If-Match"),
		req.Header.Get("If-None-Match"),
		req.Header.Get("If-Unmodified-Since"),
		req.Header.Get("Range"),
		canonicalizedHeaders(req),
		canonicalizedResource(req, account),
	}, "\n")

	mac := hmac.New(sha256.New, keyBytes)
	_, _ = mac.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("Authorization", fmt.Sprintf("SharedKey %s:%s", account, signature))
	return nil
}

// canonicalizedHeaders builds the canonicalized headers string for signing.
func canonicalizedHeaders(req *http.Request) string {
	var msHeaders []string
	for name := range req.Header {
		lower := strings.ToLower(name)
		if strings.HasPrefix(lower, "x-ms-") {
			msHeaders = append(msHeaders, lower)
		}
	}
	sort.Strings(msHeaders)

	parts := make([]string, 0, len(msHeaders))
	for _, h := range msHeaders {
		val := strings.TrimSpace(req.Header.Get(h))
		parts = append(parts, h+":"+val)
	}
	return strings.Join(parts, "\n")
}

// canonicalizedResource builds the canonicalized resource string for signing.
func canonicalizedResource(req *http.Request, account string) string {
	resource := "/" + account + req.URL.Path

	params := req.URL.Query()
	if len(params) == 0 {
		return resource
	}

	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, strings.ToLower(k))
	}
	sort.Strings(keys)

	for _, k := range keys {
		vals := params[k]
		sort.Strings(vals)
		resource += "\n" + k + ":" + strings.Join(vals, ",")
	}
	return resource
}

// resolveEndpoint extracts the endpoint, container from credentials.
func (a *AzureBlob) resolveEndpoint(creds *provider.Credentials) (string, string, error) {
	account := creds.AccessToken
	if account == "" {
		return "", "", fmt.Errorf("azure blob: missing account name in credentials")
	}

	container := ""
	if creds.Extra != nil {
		container = creds.Extra["container"]
	}
	if container == "" {
		return "", "", fmt.Errorf("azure blob: missing container in credentials")
	}

	endpoint := fmt.Sprintf("https://%s.blob.core.windows.net", account)
	if creds.Extra != nil && creds.Extra["endpoint"] != "" {
		endpoint = strings.TrimRight(creds.Extra["endpoint"], "/")
	}

	return endpoint, container, nil
}

// escapeRefPath escapes each segment of a blob ref path individually,
// preserving "/" as the path separator while encoding special characters.
func escapeRefPath(ref string) string {
	segments := strings.Split(ref, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return strings.Join(segments, "/")
}

// normalizePath ensures a prefix path ends with "/" if non-empty.
func normalizePath(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	path = strings.TrimPrefix(path, "/")
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

// XML structures for Azure Blob Storage list responses.

type listBlobsResult struct {
	XMLName    xml.Name  `xml:"EnumerationResults"`
	NextMarker string    `xml:"NextMarker"`
	Blobs      blobsList `xml:"Blobs"`
}

type blobsList struct {
	Blobs        []blobItem   `xml:"Blob"`
	BlobPrefixes []blobPrefix `xml:"BlobPrefix"`
}

type blobItem struct {
	Name       string         `xml:"Name"`
	Properties blobProperties `xml:"Properties"`
}

type blobProperties struct {
	LastModified  string `xml:"Last-Modified"`
	ContentType   string `xml:"Content-Type"`
	ContentMD5    string `xml:"Content-MD5"`
	ContentLength int64  `xml:"Content-Length"`
}

type blobPrefix struct {
	Name string `xml:"Name"`
}
