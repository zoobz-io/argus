package s3

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"sort"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*S3)(nil)

// --- test helpers ---

// fakeServer returns an httptest.Server that routes by path prefix.
// Every request is validated for the presence of an Authorization header.
func fakeServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	prefixes := make([]string, 0, len(handlers))
	for p := range handlers {
		prefixes = append(prefixes, p)
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i]) > len(prefixes[j])
	})
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Validate that requests carry AWS Signature V4 Authorization header.
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 ") {
			t.Errorf("missing or invalid Authorization header: %q", auth)
			http.Error(w, "unauthorized", http.StatusForbidden)
			return
		}

		for _, prefix := range prefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				handlers[prefix](w, r)
				return
			}
		}
		t.Errorf("unhandled request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "not found", http.StatusNotFound)
	}))
}

func xmlResponse(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/xml")
	b, err := xml.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(b)
}

// newTestProvider creates an S3 provider pointed at a test server.
func newTestProvider(t *testing.T, server *httptest.Server) *S3 {
	t.Helper()
	return New(WithEndpoint(server.URL), WithHTTPClient(server.Client()))
}

// validCreds returns test credentials with bucket and region set.
func validCreds() *provider.Credentials {
	return &provider.Credentials{
		AccessToken:  "AKIAIOSFODNN7EXAMPLE",
		RefreshToken: "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
		Extra: map[string]string{
			"region":   "us-east-1",
			"bucket":   "test-bucket",
			"endpoint": "",
		},
	}
}

// --- unit tests: Type ---

func TestType(t *testing.T) {
	s := New()
	if s.Type() != "s3" {
		t.Errorf("type: got %q, want %q", s.Type(), "s3")
	}
}

// --- unit tests: AuthURL / Exchange ---

func TestAuthURL_ReturnsError(t *testing.T) {
	s := New()
	u, err := s.AuthURL(context.Background(), "https://example.com/callback", "state")
	if err == nil {
		t.Fatal("expected error from AuthURL")
	}
	if u != "" {
		t.Errorf("expected empty URL, got %q", u)
	}
	if !strings.Contains(err.Error(), "static credentials") {
		t.Errorf("error should mention static credentials, got %q", err.Error())
	}
}

func TestExchange_ReturnsError(t *testing.T) {
	s := New()
	creds, err := s.Exchange(context.Background(), "code", "https://example.com/callback")
	if err == nil {
		t.Fatal("expected error from Exchange")
	}
	if creds != nil {
		t.Error("expected nil credentials")
	}
	if !strings.Contains(err.Error(), "static credentials") {
		t.Errorf("error should mention static credentials, got %q", err.Error())
	}
}

// --- unit tests: helpers ---

func TestBucketFromCreds_Nil(t *testing.T) {
	_, err := bucketFromCreds(nil)
	if err == nil {
		t.Fatal("expected error for nil creds")
	}
}

func TestBucketFromCreds_MissingBucket(t *testing.T) {
	_, err := bucketFromCreds(&provider.Credentials{})
	if err == nil {
		t.Fatal("expected error for missing bucket")
	}
}

func TestBucketFromCreds_Valid(t *testing.T) {
	bucket, err := bucketFromCreds(validCreds())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if bucket != "test-bucket" {
		t.Errorf("bucket: got %q", bucket)
	}
}

func TestNormalizePrefixForList(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"", ""},
		{"/", ""},
		{"docs", "docs/"},
		{"docs/", "docs/"},
		{"/docs", "docs/"},
		{"/docs/sub", "docs/sub/"},
	}
	for _, tt := range tests {
		if got := normalizePrefixForList(tt.input); got != tt.want {
			t.Errorf("normalizePrefixForList(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestPrefixName(t *testing.T) {
	tests := []struct {
		key, parent, want string
	}{
		{"docs/report.pdf", "docs/", "report.pdf"},
		{"docs/sub/", "docs/", "sub"},
		{"report.pdf", "", "report.pdf"},
	}
	for _, tt := range tests {
		if got := prefixName(tt.key, tt.parent); got != tt.want {
			t.Errorf("prefixName(%q, %q) = %q, want %q", tt.key, tt.parent, got, tt.want)
		}
	}
}

func TestObjectName(t *testing.T) {
	tests := []struct {
		key, want string
	}{
		{"docs/report.pdf", "report.pdf"},
		{"report.pdf", "report.pdf"},
		{"docs/sub/file.txt", "file.txt"},
		{"docs/sub/", "sub"},
	}
	for _, tt := range tests {
		if got := objectName(tt.key); got != tt.want {
			t.Errorf("objectName(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestEscapeObjectKey(t *testing.T) {
	tests := []struct {
		key, want string
	}{
		{"docs/report.pdf", "docs/report.pdf"},
		{"docs/my file.pdf", "docs/my%20file.pdf"},
		{"report.pdf", "report.pdf"},
		{"path/with spaces/and+plus.txt", "path/with%20spaces/and+plus.txt"},
	}
	for _, tt := range tests {
		if got := escapeObjectKey(tt.key); got != tt.want {
			t.Errorf("escapeObjectKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestResolveEndpoint_Override(t *testing.T) {
	s := New(WithEndpoint("http://localhost:9000"))
	got := s.resolveEndpoint(validCreds())
	if got != "http://localhost:9000" {
		t.Errorf("expected override endpoint, got %q", got)
	}
}

func TestResolveEndpoint_FromCreds(t *testing.T) {
	s := New()
	creds := validCreds()
	creds.Extra["endpoint"] = "http://minio.local:9000"
	got := s.resolveEndpoint(creds)
	if got != "http://minio.local:9000" {
		t.Errorf("expected creds endpoint, got %q", got)
	}
}

func TestResolveEndpoint_Default(t *testing.T) {
	s := New()
	creds := validCreds()
	creds.Extra["endpoint"] = ""
	got := s.resolveEndpoint(creds)
	if got != "https://s3.us-east-1.amazonaws.com" {
		t.Errorf("expected default AWS endpoint, got %q", got)
	}
}

func TestResolveEndpoint_DefaultRegion(t *testing.T) {
	s := New()
	creds := &provider.Credentials{
		Extra: map[string]string{"bucket": "b"},
	}
	got := s.resolveEndpoint(creds)
	if got != "https://s3.us-east-1.amazonaws.com" {
		t.Errorf("expected us-east-1 default, got %q", got)
	}
}

func TestCredentialsNeverExpire(t *testing.T) {
	creds := validCreds()
	if creds.Expired() {
		t.Error("S3 credentials with zero expiry should never be expired")
	}
}

// --- unit tests: signing ---

func TestSignRequest_SetsAuthorizationHeader(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://s3.us-east-1.amazonaws.com/bucket/key", nil)
	s := New()
	creds := validCreds()
	s.signRequest(req, creds)

	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "AWS4-HMAC-SHA256 ") {
		t.Errorf("expected AWS4-HMAC-SHA256 auth header, got %q", auth)
	}
	if !strings.Contains(auth, "Credential=AKIAIOSFODNN7EXAMPLE/") {
		t.Errorf("auth header missing access key: %q", auth)
	}
	if !strings.Contains(auth, "SignedHeaders=") {
		t.Errorf("auth header missing signed headers: %q", auth)
	}
	if !strings.Contains(auth, "Signature=") {
		t.Errorf("auth header missing signature: %q", auth)
	}

	amzDate := req.Header.Get("X-Amz-Date")
	if amzDate == "" {
		t.Error("missing X-Amz-Date header")
	}
}

func TestSignRequest_SkipsWithoutCredentials(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	s := New()

	// nil creds
	s.signRequest(req, nil)
	if req.Header.Get("Authorization") != "" {
		t.Error("should not sign with nil creds")
	}

	// empty access key
	s.signRequest(req, &provider.Credentials{})
	if req.Header.Get("Authorization") != "" {
		t.Error("should not sign with empty access key")
	}
}

// --- integration tests: List ---

func TestList_Success(t *testing.T) {
	modTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("list-type") != "2" {
				t.Error("expected list-type=2")
			}
			xmlResponse(w, listBucketResult{
				CommonPrefixes: []commonPrefix{
					{Prefix: "docs/photos/"},
				},
				Contents: []s3Object{
					{Key: "docs/report.pdf", Size: 2048, ETag: `"abc123"`, LastModified: modTime},
					{Key: "docs/notes.txt", Size: 512, ETag: `"def456"`, LastModified: modTime},
				},
			})
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	entries, updatedCr, err := s.List(context.Background(), validCreds(), "docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	// First entry should be the folder.
	if !entries[0].IsFolder || entries[0].Name != "photos" {
		t.Errorf("entry 0: expected folder 'photos', got %+v", entries[0])
	}
	// File entries.
	if entries[1].Name != "report.pdf" || entries[1].Size != 2048 {
		t.Errorf("entry 1: %+v", entries[1])
	}
	if entries[2].Name != "notes.txt" {
		t.Errorf("entry 2: %+v", entries[2])
	}
	// S3 never refreshes creds.
	if updatedCr != nil {
		t.Error("expected nil updated creds")
	}
}

func TestList_EmptyPath(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("prefix") != "" {
				t.Errorf("expected empty prefix for root, got %q", r.URL.Query().Get("prefix"))
			}
			xmlResponse(w, listBucketResult{
				Contents: []s3Object{
					{Key: "file.txt", Size: 100, ETag: `"aaa"`, LastModified: time.Now()},
				},
			})
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	entries, _, err := s.List(context.Background(), validCreds(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
}

func TestList_MissingBucket(t *testing.T) {
	s := New()
	_, _, err := s.List(context.Background(), &provider.Credentials{}, "docs")
	if err == nil {
		t.Fatal("expected error for missing bucket")
	}
	if !strings.Contains(err.Error(), "missing bucket") {
		t.Errorf("error should mention missing bucket, got %q", err.Error())
	}
}

func TestList_Pagination(t *testing.T) {
	callCount := 0
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				xmlResponse(w, listBucketResult{
					IsTruncated:           true,
					NextContinuationToken: "token-2",
					Contents: []s3Object{
						{Key: "a.txt", Size: 10, LastModified: time.Now()},
					},
				})
				return
			}
			// Verify continuation token was passed properly.
			if ct := r.URL.Query().Get("continuation-token"); ct != "token-2" {
				t.Errorf("expected continuation-token=token-2, got %q", ct)
			}
			xmlResponse(w, listBucketResult{
				Contents: []s3Object{
					{Key: "b.txt", Size: 20, LastModified: time.Now()},
				},
			})
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	entries, _, err := s.List(context.Background(), validCreds(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries across pages, got %d", len(entries))
	}
	if callCount != 2 {
		t.Errorf("expected 2 API calls, got %d", callCount)
	}
}

// --- integration tests: Changes ---

func TestChanges_InitialSync(t *testing.T) {
	modTime := time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC)

	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket": func(w http.ResponseWriter, _ *http.Request) {
			xmlResponse(w, listBucketResult{
				Contents: []s3Object{
					{Key: "docs/report.pdf", Size: 2048, ETag: `"abc"`, LastModified: modTime},
				},
			})
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	changes, token, updatedCr, err := s.Changes(context.Background(), validCreds(), "docs", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Initial sync returns all objects as created.
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Type != provider.ChangeCreated {
		t.Errorf("expected created, got %q", changes[0].Type)
	}
	if changes[0].Entry == nil || changes[0].Entry.Name != "report.pdf" {
		t.Errorf("unexpected entry: %+v", changes[0].Entry)
	}
	// Token should be a valid RFC3339 timestamp.
	if _, parseErr := time.Parse(time.RFC3339, token); parseErr != nil {
		t.Errorf("sync token should be RFC3339, got %q", token)
	}
	if updatedCr != nil {
		t.Error("expected nil updated creds")
	}
}

func TestChanges_WithSyncToken(t *testing.T) {
	oldTime := time.Date(2025, 5, 1, 0, 0, 0, 0, time.UTC)
	newTime := time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)
	syncTime := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)

	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket": func(w http.ResponseWriter, _ *http.Request) {
			xmlResponse(w, listBucketResult{
				Contents: []s3Object{
					{Key: "old.txt", Size: 100, LastModified: oldTime},
					{Key: "new.txt", Size: 200, LastModified: newTime},
				},
			})
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	changes, _, _, err := s.Changes(context.Background(), validCreds(), "", syncTime.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only new.txt should be returned (modified after sync time).
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Ref != "new.txt" {
		t.Errorf("expected new.txt, got %q", changes[0].Ref)
	}
	if changes[0].Type != provider.ChangeModified {
		t.Errorf("expected modified, got %q", changes[0].Type)
	}
}

func TestChanges_InvalidSyncToken(t *testing.T) {
	s := New()
	_, _, _, err := s.Changes(context.Background(), validCreds(), "", "not-a-timestamp")
	if err == nil {
		t.Fatal("expected error for invalid sync token")
	}
	if !strings.Contains(err.Error(), "invalid sync token") {
		t.Errorf("error should mention invalid sync token, got %q", err.Error())
	}
}

// --- integration tests: Fetch ---

func TestFetch_Success(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket/docs/report.pdf": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Length", "9")
			w.Header().Set("ETag", `"hash123"`)
			_, _ = w.Write([]byte("pdf bytes"))
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	rc, meta, updatedCr, err := s.Fetch(context.Background(), validCreds(), "docs/report.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if meta.Name != "report.pdf" {
		t.Errorf("name: got %q", meta.Name)
	}
	if meta.MimeType != "application/pdf" {
		t.Errorf("mime: got %q", meta.MimeType)
	}
	if meta.ContentHash != `"hash123"` {
		t.Errorf("hash: got %q", meta.ContentHash)
	}
	if meta.Size != 9 {
		t.Errorf("size: got %d", meta.Size)
	}

	body, _ := io.ReadAll(rc)
	if string(body) != "pdf bytes" {
		t.Errorf("content: got %q", string(body))
	}
	if updatedCr != nil {
		t.Error("expected nil updated creds")
	}
}

func TestFetch_EmptyRef(t *testing.T) {
	s := New()
	_, _, _, err := s.Fetch(context.Background(), validCreds(), "")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
	if !strings.Contains(err.Error(), "empty object key") {
		t.Errorf("error should mention empty key, got %q", err.Error())
	}
}

func TestFetch_NotFound(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket/missing.txt": func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	_, _, _, err := s.Fetch(context.Background(), validCreds(), "missing.txt")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("error should mention 404, got %q", err.Error())
	}
}

func TestFetch_MissingBucket(t *testing.T) {
	s := New()
	_, _, _, err := s.Fetch(context.Background(), &provider.Credentials{}, "key")
	if err == nil {
		t.Fatal("expected error for missing bucket")
	}
}

// --- integration tests: Fetch with special characters (url.PathEscape) ---

func TestFetch_SpecialCharactersInKey(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket/docs/my file.pdf": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Length", "4")
			_, _ = w.Write([]byte("data"))
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	rc, meta, _, err := s.Fetch(context.Background(), validCreds(), "docs/my file.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if meta.Name != "my file.pdf" {
		t.Errorf("name: got %q, want %q", meta.Name, "my file.pdf")
	}
}

// --- integration tests: query parameter encoding ---

func TestList_QueryParamsProperlyEncoded(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/test-bucket": func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query()
			if q.Get("list-type") != "2" {
				t.Errorf("list-type: got %q", q.Get("list-type"))
			}
			if q.Get("prefix") != "path with spaces/" {
				t.Errorf("prefix: got %q, want %q", q.Get("prefix"), "path with spaces/")
			}
			if q.Get("delimiter") != "/" {
				t.Errorf("delimiter: got %q", q.Get("delimiter"))
			}
			xmlResponse(w, listBucketResult{})
		},
	})
	defer server.Close()

	s := newTestProvider(t, server)
	_, _, err := s.List(context.Background(), validCreds(), "path with spaces")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
