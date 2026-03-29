// Package s3 implements the provider interface for Amazon S3 and S3-compatible
// object stores (MinIO, DigitalOcean Spaces, etc.).
//
// S3 uses static credentials (access key + secret key) instead of OAuth2.
// The AuthURL and Exchange methods return errors — callers should supply
// credentials directly via the Credentials struct.
//
// Credential mapping:
//
//	AccessToken  = AWS access key ID
//	RefreshToken = AWS secret access key
//	Extra["region"]   = AWS region
//	Extra["bucket"]   = S3 bucket name
//	Extra["endpoint"] = optional custom endpoint (MinIO, etc.)
package s3

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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

const providerType = "s3"

// errNotOAuth is returned by AuthURL and Exchange. S3 uses static credentials.
var errNotOAuth = fmt.Errorf("s3 uses static credentials, not OAuth")

// S3 implements provider.Provider for Amazon S3 and compatible stores.
type S3 struct {
	// httpClient is the HTTP client used for S3 API requests.
	// Defaults to http.DefaultClient when nil.
	httpClient *http.Client

	// endpointOverride forces all requests to a specific base URL.
	// Used for testing (httptest.Server) and S3-compatible stores.
	endpointOverride string
}

// Option configures the S3 provider.
type Option func(*S3)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) Option {
	return func(s *S3) { s.httpClient = c }
}

// WithEndpoint overrides the S3 endpoint. Used for testing and
// S3-compatible stores like MinIO.
func WithEndpoint(endpoint string) Option {
	return func(s *S3) { s.endpointOverride = endpoint }
}

// New creates an S3 provider with the given options.
func New(opts ...Option) *S3 {
	s := &S3{}
	for _, o := range opts {
		o(s)
	}
	return s
}

// Type returns the provider type identifier.
func (s *S3) Type() string {
	return providerType
}

// AuthURL returns an error. S3 uses static credentials, not OAuth.
func (s *S3) AuthURL(_ context.Context, _, _ string) (string, error) {
	return "", errNotOAuth
}

// Exchange returns an error. S3 uses static credentials, not OAuth.
func (s *S3) Exchange(_ context.Context, _, _ string) (*provider.Credentials, error) {
	return nil, errNotOAuth
}

// List returns entries (objects and common prefixes) at the given key prefix.
// The path is treated as a prefix with "/" as delimiter for folder-like browsing.
func (s *S3) List(ctx context.Context, creds *provider.Credentials, path string) ([]provider.Entry, *provider.Credentials, error) {
	bucket, err := bucketFromCreds(creds)
	if err != nil {
		return nil, nil, err
	}

	prefix := normalizePrefixForList(path)

	var entries []provider.Entry
	var continuationToken string

	for {
		resp, listErr := s.listObjectsV2(ctx, creds, bucket, prefix, "/", continuationToken)
		if listErr != nil {
			return nil, nil, fmt.Errorf("listing objects: %w", listErr)
		}

		for _, cp := range resp.CommonPrefixes {
			name := prefixName(cp.Prefix, prefix)
			if name == "" {
				continue
			}
			entries = append(entries, provider.Entry{
				Ref:      cp.Prefix,
				Name:     name,
				IsFolder: true,
			})
		}

		for _, obj := range resp.Contents {
			// Skip the prefix itself if it shows up as an object.
			if obj.Key == prefix {
				continue
			}
			name := prefixName(obj.Key, prefix)
			if name == "" {
				continue
			}
			entries = append(entries, provider.Entry{
				Ref:         obj.Key,
				Name:        name,
				Size:        obj.Size,
				ContentHash: obj.ETag,
				ModifiedAt:  obj.LastModified,
			})
		}

		if !resp.IsTruncated {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	return entries, nil, nil
}

// Changes returns entries modified since the given sync token.
// The sync token is an RFC3339 timestamp of the last sync.
// Pass empty token for initial sync (returns all objects as created).
func (s *S3) Changes(ctx context.Context, creds *provider.Credentials, path, syncToken string) ([]provider.Change, string, *provider.Credentials, error) {
	bucket, err := bucketFromCreds(creds)
	if err != nil {
		return nil, syncToken, nil, err
	}

	prefix := normalizePrefixForList(path)
	now := time.Now().UTC()

	var since time.Time
	if syncToken != "" {
		since, err = time.Parse(time.RFC3339, syncToken)
		if err != nil {
			return nil, syncToken, nil, fmt.Errorf("invalid sync token: %w", err)
		}
	}

	var changes []provider.Change
	var continuationToken string

	for {
		resp, listErr := s.listObjectsV2(ctx, creds, bucket, prefix, "", continuationToken)
		if listErr != nil {
			return nil, syncToken, nil, fmt.Errorf("listing changes: %w", listErr)
		}

		for _, obj := range resp.Contents {
			if !since.IsZero() && !obj.LastModified.After(since) {
				continue
			}

			entry := provider.Entry{
				Ref:         obj.Key,
				Name:        objectName(obj.Key),
				Size:        obj.Size,
				ContentHash: obj.ETag,
				ModifiedAt:  obj.LastModified,
			}

			changeType := provider.ChangeModified
			if since.IsZero() {
				changeType = provider.ChangeCreated
			}

			changes = append(changes, provider.Change{
				Ref:   obj.Key,
				Entry: &entry,
				Type:  changeType,
			})
		}

		if !resp.IsTruncated {
			break
		}
		continuationToken = resp.NextContinuationToken
	}

	return changes, now.Format(time.RFC3339), nil, nil
}

// Fetch downloads an object by key.
// The caller must close the returned ReadCloser.
func (s *S3) Fetch(ctx context.Context, creds *provider.Credentials, ref string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	bucket, err := bucketFromCreds(creds)
	if err != nil {
		return nil, nil, nil, err
	}

	if ref == "" {
		return nil, nil, nil, fmt.Errorf("empty object key")
	}

	endpoint := s.resolveEndpoint(creds)
	escapedRef := escapeObjectKey(ref)
	reqURL := fmt.Sprintf("%s/%s/%s", endpoint, bucket, escapedRef)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("creating request: %w", err)
	}

	s.signRequest(req, creds)

	resp, err := s.client().Do(req)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("fetching object: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, nil, nil, fmt.Errorf("fetching object: status %d", resp.StatusCode)
	}

	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
	meta := &provider.EntryMeta{
		Name:        objectName(ref),
		MimeType:    resp.Header.Get("Content-Type"),
		ContentHash: resp.Header.Get("ETag"),
		Size:        size,
	}

	return resp.Body, meta, nil, nil
}

// --- S3 XML types ---

type listBucketResult struct {
	XMLName               xml.Name       `xml:"ListBucketResult"`
	NextContinuationToken string         `xml:"NextContinuationToken"`
	Contents              []s3Object     `xml:"Contents"`
	CommonPrefixes        []commonPrefix `xml:"CommonPrefixes"`
	IsTruncated           bool           `xml:"IsTruncated"`
}

type s3Object struct {
	LastModified time.Time `xml:"LastModified"`
	Key          string    `xml:"Key"`
	ETag         string    `xml:"ETag"`
	Size         int64     `xml:"Size"`
}

type commonPrefix struct {
	Prefix string `xml:"Prefix"`
}

// --- internal helpers ---

// listObjectsV2 calls the S3 ListObjectsV2 API.
func (s *S3) listObjectsV2(ctx context.Context, creds *provider.Credentials, bucket, prefix, delimiter, continuationToken string) (*listBucketResult, error) {
	endpoint := s.resolveEndpoint(creds)

	params := url.Values{}
	params.Set("list-type", "2")
	if prefix != "" {
		params.Set("prefix", prefix)
	}
	if delimiter != "" {
		params.Set("delimiter", delimiter)
	}
	if continuationToken != "" {
		params.Set("continuation-token", continuationToken)
	}

	reqURL := fmt.Sprintf("%s/%s?%s", endpoint, bucket, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	s.signRequest(req, creds)

	resp, err := s.client().Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	var result listBucketResult
	if err := xml.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decoding response: %w", err)
	}

	return &result, nil
}

// signRequest applies AWS Signature V4 to an outbound HTTP request.
// If credentials are missing access key or secret key, the request is
// sent unsigned (allows test servers to work without signing).
func (s *S3) signRequest(req *http.Request, creds *provider.Credentials) {
	if creds == nil || creds.AccessToken == "" || creds.RefreshToken == "" {
		return
	}

	accessKey := creds.AccessToken
	secretKey := creds.RefreshToken
	region := "us-east-1"
	if creds.Extra != nil && creds.Extra["region"] != "" {
		region = creds.Extra["region"]
	}

	now := time.Now().UTC()
	datestamp := now.Format("20060102")
	amzDate := now.Format("20060102T150405Z")
	service := "s3"

	// Set required headers before building the canonical request.
	req.Header.Set("X-Amz-Date", amzDate)
	req.Header.Set("X-Amz-Content-Sha256", "UNSIGNED-PAYLOAD")
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}

	// Canonical request components.
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Canonical query string: sorted by key.
	canonicalQueryString := canonicalQuery(req.URL.Query())

	// Signed headers: sorted lowercase header names.
	signedHeaders, canonicalHeaders := canonicalHeaderString(req)

	payloadHash := "UNSIGNED-PAYLOAD"

	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	// String to sign.
	credentialScope := fmt.Sprintf("%s/%s/%s/aws4_request", datestamp, region, service)
	stringToSign := strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		sha256Hex([]byte(canonicalRequest)),
	}, "\n")

	// Signing key.
	signingKey := deriveSigningKey(secretKey, datestamp, region, service)

	// Signature.
	signature := hex.EncodeToString(hmacSHA256(signingKey, []byte(stringToSign)))

	// Authorization header.
	authHeader := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		accessKey, credentialScope, signedHeaders, signature)
	req.Header.Set("Authorization", authHeader)
}

// deriveSigningKey produces the AWS Signature V4 signing key.
func deriveSigningKey(secret, datestamp, region, service string) []byte {
	kDate := hmacSHA256([]byte("AWS4"+secret), []byte(datestamp))
	kRegion := hmacSHA256(kDate, []byte(region))
	kService := hmacSHA256(kRegion, []byte(service))
	return hmacSHA256(kService, []byte("aws4_request"))
}

// hmacSHA256 computes HMAC-SHA256.
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// sha256Hex returns the lowercase hex SHA-256 digest.
func sha256Hex(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}

// canonicalQuery builds the canonical query string (sorted by key, then value).
func canonicalQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		vs := values[k]
		sort.Strings(vs)
		for _, v := range vs {
			parts = append(parts, url.QueryEscape(k)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(parts, "&")
}

// canonicalHeaderString returns the signed-header list and the canonical
// header block for AWS Signature V4.
func canonicalHeaderString(req *http.Request) (signedHeaders, canonicalHeaders string) {
	headers := map[string]string{}
	names := make([]string, 0, len(req.Header))
	for name := range req.Header {
		lc := strings.ToLower(name)
		headers[lc] = strings.TrimSpace(req.Header.Get(name))
		names = append(names, lc)
	}
	sort.Strings(names)

	headerLines := make([]string, 0, len(names))
	for _, n := range names {
		headerLines = append(headerLines, n+":"+headers[n])
	}
	canonicalHeaders = strings.Join(headerLines, "\n") + "\n"
	signedHeaders = strings.Join(names, ";")
	return signedHeaders, canonicalHeaders
}

// escapeObjectKey percent-encodes each path segment of an S3 object key,
// preserving "/" separators.
func escapeObjectKey(key string) string {
	segments := strings.Split(key, "/")
	for i, seg := range segments {
		segments[i] = url.PathEscape(seg)
	}
	return strings.Join(segments, "/")
}

// resolveEndpoint returns the S3 endpoint URL.
func (s *S3) resolveEndpoint(creds *provider.Credentials) string {
	if s.endpointOverride != "" {
		return s.endpointOverride
	}
	if creds.Extra != nil && creds.Extra["endpoint"] != "" {
		return strings.TrimRight(creds.Extra["endpoint"], "/")
	}
	region := "us-east-1"
	if creds.Extra != nil && creds.Extra["region"] != "" {
		region = creds.Extra["region"]
	}
	return fmt.Sprintf("https://s3.%s.amazonaws.com", region)
}

// client returns the HTTP client to use for requests.
func (s *S3) client() *http.Client {
	if s.httpClient != nil {
		return s.httpClient
	}
	return http.DefaultClient
}

// bucketFromCreds extracts and validates the bucket name from credentials.
func bucketFromCreds(creds *provider.Credentials) (string, error) {
	if creds == nil {
		return "", fmt.Errorf("nil credentials")
	}
	if creds.Extra == nil || creds.Extra["bucket"] == "" {
		return "", fmt.Errorf("missing bucket in credentials")
	}
	return creds.Extra["bucket"], nil
}

// normalizePrefixForList ensures the prefix ends with "/" for folder-like browsing.
// Empty and root paths return empty string (list all).
func normalizePrefixForList(path string) string {
	if path == "" || path == "/" {
		return ""
	}
	path = strings.TrimPrefix(path, "/")
	if !strings.HasSuffix(path, "/") {
		path += "/"
	}
	return path
}

// prefixName extracts the display name from a key relative to the parent prefix.
func prefixName(key, parentPrefix string) string {
	name := strings.TrimPrefix(key, parentPrefix)
	name = strings.TrimSuffix(name, "/")
	return name
}

// objectName extracts the filename from an object key.
func objectName(key string) string {
	key = strings.TrimSuffix(key, "/")
	if idx := strings.LastIndex(key, "/"); idx >= 0 {
		return key[idx+1:]
	}
	return key
}
