package azureblob

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*AzureBlob)(nil)

// --- test helpers ---

func validCreds(endpoint string) *provider.Credentials {
	return &provider.Credentials{
		AccessToken:  "testaccount",
		RefreshToken: "dGVzdGtleQ==", // base64("testkey")
		Extra: map[string]string{
			"container": "testcontainer",
			"endpoint":  endpoint,
		},
	}
}

// --- unit tests: Type ---

func TestType(t *testing.T) {
	a := New()
	if a.Type() != "azure_blob" {
		t.Errorf("type: got %q, want %q", a.Type(), "azure_blob")
	}
}

// --- unit tests: AuthURL / Exchange ---

func TestAuthURL_ReturnsError(t *testing.T) {
	a := New()
	_, err := a.AuthURL(context.Background(), "https://app.example.com/callback", "state")
	if err == nil {
		t.Fatal("expected error from AuthURL")
	}
	if !strings.Contains(err.Error(), "static credentials") {
		t.Errorf("error should mention static credentials, got %q", err.Error())
	}
}

func TestExchange_ReturnsError(t *testing.T) {
	a := New()
	_, err := a.Exchange(context.Background(), "code", "https://app.example.com/callback")
	if err == nil {
		t.Fatal("expected error from Exchange")
	}
	if !strings.Contains(err.Error(), "static credentials") {
		t.Errorf("error should mention static credentials, got %q", err.Error())
	}
}

// --- unit tests: helpers ---

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"", ""},
		{"/", ""},
		{"docs", "docs/"},
		{"docs/", "docs/"},
		{"/docs", "docs/"},
		{"/docs/sub", "docs/sub/"},
	}
	for _, tt := range tests {
		if got := normalizePath(tt.input); got != tt.expected {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestResolveEndpoint_Default(t *testing.T) {
	creds := &provider.Credentials{
		AccessToken: "myaccount",
		Extra:       map[string]string{"container": "mycontainer"},
	}
	endpoint, container, err := New().resolveEndpoint(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != "https://myaccount.blob.core.windows.net" {
		t.Errorf("endpoint: got %q", endpoint)
	}
	if container != "mycontainer" {
		t.Errorf("container: got %q", container)
	}
}

func TestResolveEndpoint_Custom(t *testing.T) {
	creds := &provider.Credentials{
		AccessToken: "myaccount",
		Extra:       map[string]string{"container": "mycontainer", "endpoint": "http://localhost:10000/devstoreaccount1"},
	}
	endpoint, _, err := New().resolveEndpoint(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != "http://localhost:10000/devstoreaccount1" {
		t.Errorf("endpoint: got %q", endpoint)
	}
}

func TestResolveEndpoint_MissingAccount(t *testing.T) {
	creds := &provider.Credentials{
		Extra: map[string]string{"container": "c"},
	}
	_, _, err := New().resolveEndpoint(creds)
	if err == nil {
		t.Fatal("expected error for missing account")
	}
}

func TestResolveEndpoint_MissingContainer(t *testing.T) {
	creds := &provider.Credentials{
		AccessToken: "acct",
	}
	_, _, err := New().resolveEndpoint(creds)
	if err == nil {
		t.Fatal("expected error for missing container")
	}
}

func TestCanonicalizedHeaders(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	req.Header.Set("x-ms-date", "Mon, 01 Jan 2024 00:00:00 GMT")
	req.Header.Set("x-ms-version", "2020-10-02")
	req.Header.Set("Other-Header", "ignored")

	result := canonicalizedHeaders(req)
	if !strings.Contains(result, "x-ms-date:") {
		t.Errorf("should contain x-ms-date, got %q", result)
	}
	if !strings.Contains(result, "x-ms-version:") {
		t.Errorf("should contain x-ms-version, got %q", result)
	}
	if strings.Contains(result, "other-header") {
		t.Errorf("should not contain non-ms headers, got %q", result)
	}
	// Verify ordering.
	dateIdx := strings.Index(result, "x-ms-date:")
	versionIdx := strings.Index(result, "x-ms-version:")
	if dateIdx > versionIdx {
		t.Errorf("headers should be sorted, got %q", result)
	}
}

func TestCanonicalizedResource(t *testing.T) {
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://acct.blob.core.windows.net/container?restype=container&comp=list", nil)
	result := canonicalizedResource(req, "acct")
	if !strings.HasPrefix(result, "/acct/container") {
		t.Errorf("should start with /acct/container, got %q", result)
	}
	if !strings.Contains(result, "comp:list") {
		t.Errorf("should contain comp:list, got %q", result)
	}
	if !strings.Contains(result, "restype:container") {
		t.Errorf("should contain restype:container, got %q", result)
	}
}

// --- integration tests: full method calls through mock HTTP ---

func TestList_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("comp") != "list" {
			t.Errorf("expected comp=list, got %q", r.URL.Query().Get("comp"))
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <BlobPrefix><Name>docs/</Name></BlobPrefix>
    <Blob>
      <Name>report.pdf</Name>
      <Properties>
        <Last-Modified>Sat, 01 Jun 2025 12:00:00 GMT</Last-Modified>
        <Content-Type>application/pdf</Content-Type>
        <Content-Length>2048</Content-Length>
        <Content-MD5>abc123</Content-MD5>
      </Properties>
    </Blob>
  </Blobs>
</EnumerationResults>`))
	}))
	defer server.Close()

	a := New()
	entries, updatedCr, err := a.List(context.Background(), validCreds(server.URL), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	// First entry is the virtual directory.
	if entries[0].Name != "docs" || !entries[0].IsFolder {
		t.Errorf("entry 0: got %+v", entries[0])
	}
	// Second entry is the blob.
	if entries[1].Name != "report.pdf" || entries[1].Size != 2048 {
		t.Errorf("entry 1: got %+v", entries[1])
	}
	if entries[1].ContentHash != "abc123" {
		t.Errorf("entry 1 hash: got %q", entries[1].ContentHash)
	}
	// No credential refresh for static keys.
	if updatedCr != nil {
		t.Error("expected nil updated creds")
	}
}

func TestList_WithPrefix(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		prefix := r.URL.Query().Get("prefix")
		if prefix != "docs/" {
			t.Errorf("expected prefix=docs/, got %q", prefix)
		}
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults><Blobs></Blobs></EnumerationResults>`))
	}))
	defer server.Close()

	a := New()
	entries, _, err := a.List(context.Background(), validCreds(server.URL), "docs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestList_MissingContainer(t *testing.T) {
	a := New()
	creds := &provider.Credentials{
		AccessToken:  "acct",
		RefreshToken: "dGVzdGtleQ==",
	}
	_, _, err := a.List(context.Background(), creds, "")
	if err == nil {
		t.Fatal("expected error for missing container")
	}
	if !strings.Contains(err.Error(), "missing container") {
		t.Errorf("error should mention missing container, got %q", err.Error())
	}
}

func TestList_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	a := New()
	_, _, err := a.List(context.Background(), validCreds(server.URL), "")
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("error should contain status code, got %q", err.Error())
	}
}

func TestList_Pagination(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/xml")
		if n == 1 {
			// First page: one blob + NextMarker.
			if r.URL.Query().Get("marker") != "" {
				t.Errorf("first request should not have marker, got %q", r.URL.Query().Get("marker"))
			}
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <Blob>
      <Name>file1.txt</Name>
      <Properties>
        <Content-Type>text/plain</Content-Type>
        <Content-Length>10</Content-Length>
      </Properties>
    </Blob>
  </Blobs>
  <NextMarker>marker-page2</NextMarker>
</EnumerationResults>`))
			return
		}
		// Second page: one blob, no NextMarker.
		if r.URL.Query().Get("marker") != "marker-page2" {
			t.Errorf("second request should have marker=marker-page2, got %q", r.URL.Query().Get("marker"))
		}
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <Blob>
      <Name>file2.txt</Name>
      <Properties>
        <Content-Type>text/plain</Content-Type>
        <Content-Length>20</Content-Length>
      </Properties>
    </Blob>
  </Blobs>
</EnumerationResults>`))
	}))
	defer server.Close()

	a := New()
	entries, _, err := a.List(context.Background(), validCreds(server.URL), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries across pages, got %d", len(entries))
	}
	if entries[0].Name != "file1.txt" {
		t.Errorf("entry 0: got %q", entries[0].Name)
	}
	if entries[1].Name != "file2.txt" {
		t.Errorf("entry 1: got %q", entries[1].Name)
	}
	if calls.Load() != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", calls.Load())
	}
}

func TestChanges_InitialSync(t *testing.T) {
	a := New()
	changes, token, updatedCr, err := a.Changes(context.Background(), validCreds("http://unused"), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes on initial sync, got %d", len(changes))
	}
	if token == "" {
		t.Error("expected non-empty sync token")
	}
	// Token should be parseable as RFC3339.
	if _, parseErr := time.Parse(time.RFC3339, token); parseErr != nil {
		t.Errorf("sync token should be RFC3339, got %q: %v", token, parseErr)
	}
	if updatedCr != nil {
		t.Error("expected nil updated creds")
	}
}

func TestChanges_WithDelta(t *testing.T) {
	// Blobs: one modified recently, one old.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <Blob>
      <Name>new-file.txt</Name>
      <Properties>
        <Last-Modified>Sat, 01 Mar 2025 12:00:00 GMT</Last-Modified>
        <Content-Type>text/plain</Content-Type>
        <Content-Length>100</Content-Length>
      </Properties>
    </Blob>
    <Blob>
      <Name>old-file.txt</Name>
      <Properties>
        <Last-Modified>Sat, 01 Jan 2025 12:00:00 GMT</Last-Modified>
        <Content-Type>text/plain</Content-Type>
        <Content-Length>200</Content-Length>
      </Properties>
    </Blob>
  </Blobs>
</EnumerationResults>`))
	}))
	defer server.Close()

	a := New()
	// Last-sync timestamp is after old-file but before new-file.
	lastSync := "2025-02-01T00:00:00Z"
	changes, newToken, _, err := a.Changes(context.Background(), validCreds(server.URL), "", lastSync)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].Ref != "new-file.txt" {
		t.Errorf("change ref: got %q", changes[0].Ref)
	}
	if changes[0].Type != provider.ChangeModified {
		t.Errorf("change type: got %q", changes[0].Type)
	}
	if newToken == "" || newToken == lastSync {
		t.Errorf("expected new sync token, got %q", newToken)
	}
}

func TestChanges_InvalidToken(t *testing.T) {
	a := New()
	_, _, _, err := a.Changes(context.Background(), validCreds("http://unused"), "", "not-a-timestamp")
	if err == nil {
		t.Fatal("expected error for invalid sync token")
	}
	if !strings.Contains(err.Error(), "parsing sync token") {
		t.Errorf("error should mention parsing, got %q", err.Error())
	}
}

func TestFetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/testcontainer/docs/report.pdf") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Length", "9")
		w.Header().Set("Content-MD5", "hash123")
		_, _ = w.Write([]byte("pdf bytes"))
	}))
	defer server.Close()

	a := New()
	rc, meta, updatedCr, err := a.Fetch(context.Background(), validCreds(server.URL), "docs/report.pdf")
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
	if meta.ContentHash != "hash123" {
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

func TestFetch_EscapesRef(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The ref "docs/my file.txt" should arrive with space encoded.
		if !strings.Contains(r.URL.RawPath, "my%20file.txt") && !strings.Contains(r.URL.RequestURI(), "my%20file.txt") {
			t.Errorf("expected escaped ref in URL, got path=%q raw=%q", r.URL.Path, r.URL.RawPath)
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Content-Length", "5")
		_, _ = w.Write([]byte("hello"))
	}))
	defer server.Close()

	a := New()
	rc, meta, _, err := a.Fetch(context.Background(), validCreds(server.URL), "docs/my file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if meta.Name != "my file.txt" {
		t.Errorf("name: got %q", meta.Name)
	}
}

func TestFetch_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "blob not found", http.StatusNotFound)
	}))
	defer server.Close()

	a := New()
	_, _, _, err := a.Fetch(context.Background(), validCreds(server.URL), "missing.txt")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "404") {
		t.Errorf("error should contain 404, got %q", err.Error())
	}
}

// --- error path tests ---

func TestList_InvalidXMLResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte("not xml"))
	}))
	defer server.Close()

	a := New()
	_, _, err := a.List(context.Background(), validCreds(server.URL), "")
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
	if !strings.Contains(err.Error(), "decoding list response") {
		t.Errorf("error should mention decoding, got %q", err.Error())
	}
}

func TestChanges_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "service unavailable", http.StatusServiceUnavailable)
	}))
	defer server.Close()

	a := New()
	_, _, _, err := a.Changes(context.Background(), validCreds(server.URL), "", "2025-02-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected error for server error")
	}
	if !strings.Contains(err.Error(), "503") {
		t.Errorf("error should mention 503, got %q", err.Error())
	}
}

func TestChanges_InvalidXMLResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		_, _ = w.Write([]byte("not xml"))
	}))
	defer server.Close()

	a := New()
	_, _, _, err := a.Changes(context.Background(), validCreds(server.URL), "", "2025-02-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected error for invalid XML")
	}
	if !strings.Contains(err.Error(), "decoding changes response") {
		t.Errorf("error should mention decoding, got %q", err.Error())
	}
}

func TestChanges_MissingContainer(t *testing.T) {
	a := New()
	creds := &provider.Credentials{
		AccessToken:  "acct",
		RefreshToken: "dGVzdGtleQ==",
	}
	_, _, _, err := a.Changes(context.Background(), creds, "", "2025-02-01T00:00:00Z")
	if err == nil {
		t.Fatal("expected error for missing container")
	}
	if !strings.Contains(err.Error(), "missing container") {
		t.Errorf("error should mention missing container, got %q", err.Error())
	}
}

func TestFetch_MissingContainer(t *testing.T) {
	a := New()
	creds := &provider.Credentials{
		AccessToken:  "acct",
		RefreshToken: "dGVzdGtleQ==",
	}
	_, _, _, err := a.Fetch(context.Background(), creds, "blob.txt")
	if err == nil {
		t.Fatal("expected error for missing container")
	}
	if !strings.Contains(err.Error(), "missing container") {
		t.Errorf("error should mention missing container, got %q", err.Error())
	}
}

func TestFetch_MissingAccount(t *testing.T) {
	a := New()
	creds := &provider.Credentials{
		Extra: map[string]string{"container": "c"},
	}
	_, _, _, err := a.Fetch(context.Background(), creds, "blob.txt")
	if err == nil {
		t.Fatal("expected error for missing account")
	}
	if !strings.Contains(err.Error(), "missing account") {
		t.Errorf("error should mention missing account, got %q", err.Error())
	}
}

func TestChanges_Pagination(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		w.Header().Set("Content-Type", "application/xml")
		if n == 1 {
			_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <Blob>
      <Name>new1.txt</Name>
      <Properties>
        <Last-Modified>Sat, 01 Mar 2025 12:00:00 GMT</Last-Modified>
        <Content-Type>text/plain</Content-Type>
        <Content-Length>100</Content-Length>
      </Properties>
    </Blob>
  </Blobs>
  <NextMarker>marker-2</NextMarker>
</EnumerationResults>`))
			return
		}
		_, _ = w.Write([]byte(`<?xml version="1.0" encoding="utf-8"?>
<EnumerationResults>
  <Blobs>
    <Blob>
      <Name>new2.txt</Name>
      <Properties>
        <Last-Modified>Sat, 01 Mar 2025 14:00:00 GMT</Last-Modified>
        <Content-Type>text/plain</Content-Type>
        <Content-Length>200</Content-Length>
      </Properties>
    </Blob>
  </Blobs>
</EnumerationResults>`))
	}))
	defer server.Close()

	a := New()
	changes, _, _, err := a.Changes(context.Background(), validCreds(server.URL), "", "2025-02-01T00:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes across pages, got %d", len(changes))
	}
	if calls.Load() != 2 {
		t.Errorf("expected 2 HTTP calls, got %d", calls.Load())
	}
}

func TestEscapeRefPath(t *testing.T) {
	tests := []struct{ input, want string }{
		{"docs/report.pdf", "docs/report.pdf"},
		{"docs/my file.pdf", "docs/my%20file.pdf"},
		{"report.pdf", "report.pdf"},
	}
	for _, tt := range tests {
		if got := escapeRefPath(tt.input); got != tt.want {
			t.Errorf("escapeRefPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestResolveEndpoint_TrailingSlash(t *testing.T) {
	creds := &provider.Credentials{
		AccessToken: "myaccount",
		Extra: map[string]string{
			"container": "c",
			"endpoint":  "http://localhost:10000/devstoreaccount1/",
		},
	}
	endpoint, _, err := New().resolveEndpoint(creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.HasSuffix(endpoint, "/") {
		t.Errorf("endpoint should not have trailing slash, got %q", endpoint)
	}
}

func TestSignRequest_InvalidKey(t *testing.T) {
	a := New()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com", nil)
	creds := &provider.Credentials{
		AccessToken:  "acct",
		RefreshToken: "not-base64!@#$",
	}
	err := a.signRequest(req, creds)
	if err == nil {
		t.Fatal("expected error for invalid base64 key")
	}
}

func TestSignRequest_SetsHeaders(t *testing.T) {
	a := New()
	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://acct.blob.core.windows.net/container?comp=list", nil)
	creds := &provider.Credentials{
		AccessToken:  "acct",
		RefreshToken: "dGVzdGtleQ==",
	}
	err := a.signRequest(req, creds)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	auth := req.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "SharedKey acct:") {
		t.Errorf("Authorization header: got %q", auth)
	}
	if req.Header.Get("x-ms-date") == "" {
		t.Error("expected x-ms-date header")
	}
	if req.Header.Get("x-ms-version") == "" {
		t.Error("expected x-ms-version header")
	}
}
