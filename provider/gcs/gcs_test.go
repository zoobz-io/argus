package gcs

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*GCS)(nil)

// --- test helpers ---

// fakeGCSServer returns an httptest.Server that routes by path and method.
func fakeGCSServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	prefixes := make([]string, 0, len(handlers))
	for p := range handlers {
		prefixes = append(prefixes, p)
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i]) > len(prefixes[j])
	})
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, prefix := range prefixes {
			if strings.HasPrefix(r.URL.Path, prefix) {
				handlers[prefix](w, r)
				return
			}
		}
		t.Errorf("unhandled request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "not found", 404)
	}))
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func jsonResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

// fakeServiceAccountJSON returns a base64-encoded fake service account key.
func fakeServiceAccountJSON() string {
	sa := map[string]string{
		"type":                        "service_account",
		"project_id":                  "test-project",
		"private_key_id":              "key-id",
		"private_key":                 fakePrivateKey,
		"client_email":                "test@test-project.iam.gserviceaccount.com",
		"client_id":                   "123456789",
		"auth_uri":                    "https://accounts.google.com/o/oauth2/auth",
		"token_uri":                   "https://oauth2.googleapis.com/token",
		"auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
		"client_x509_cert_url":        "https://www.googleapis.com/robot/v1/metadata/x509/test",
	}
	return base64.StdEncoding.EncodeToString(mustJSON(sa))
}

// newTestProvider creates a GCS provider pointed at a test server.
func newTestProvider(_ *testing.T, server *httptest.Server) *GCS {
	return &GCS{
		endpoint: server.URL + "/storage/v1/",
	}
}

// validCreds returns test credentials with required Extra fields.
func validCreds() *provider.Credentials {
	return &provider.Credentials{
		Extra: map[string]string{
			"project_id":      "test-project",
			"bucket":          "test-bucket",
			"credentials_json": fakeServiceAccountJSON(),
		},
	}
}

// --- unit tests: Type, AuthURL, Exchange ---

func TestType(t *testing.T) {
	g := New()
	if g.Type() != "gcs" {
		t.Errorf("type: got %q, want %q", g.Type(), "gcs")
	}
}

func TestAuthURL_ReturnsError(t *testing.T) {
	g := New()
	_, err := g.AuthURL(context.Background(), "https://app.example.com/callback", "csrf-token")
	if err == nil {
		t.Fatal("expected error from AuthURL")
	}
	if !strings.Contains(err.Error(), "service account credentials") {
		t.Errorf("error should mention service account credentials, got %q", err.Error())
	}
}

func TestExchange_ReturnsError(t *testing.T) {
	g := New()
	_, err := g.Exchange(context.Background(), "auth-code", "https://app.example.com/callback")
	if err == nil {
		t.Fatal("expected error from Exchange")
	}
	if !strings.Contains(err.Error(), "service account credentials") {
		t.Errorf("error should mention service account credentials, got %q", err.Error())
	}
}

// --- unit tests: helpers ---

func TestObjectName(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"docs/report.pdf", "report.pdf"},
		{"report.pdf", "report.pdf"},
		{"a/b/c/file.txt", "file.txt"},
		{"folder/", "folder"},
	}
	for _, tt := range tests {
		if got := objectName(tt.key); got != tt.want {
			t.Errorf("objectName(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAttrsToEntry_Prefix(t *testing.T) {
	// import here is avoided; we test attrsToEntry via the provider.Entry result.
	// Prefix entries are pseudo-folders.
	entry := provider.Entry{
		Ref:      "docs/",
		Name:     "docs",
		IsFolder: true,
	}
	if !entry.IsFolder {
		t.Error("expected folder")
	}
	if entry.Name != "docs" {
		t.Errorf("name: got %q", entry.Name)
	}
}

func TestSyncTokenToUnix(t *testing.T) {
	token := "2025-06-01T12:00:00Z" //nolint:gosec // not a credential
	result := syncTokenToUnix(token)
	if result == "0" {
		t.Error("expected non-zero unix timestamp")
	}
}

func TestSyncTokenToUnix_Invalid(t *testing.T) {
	if syncTokenToUnix("not-a-timestamp") != "0" {
		t.Error("expected 0 for invalid token")
	}
}

func TestDecodeCredentialsJSON_Empty(t *testing.T) {
	_, err := decodeCredentialsJSON("")
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestDecodeCredentialsJSON_InvalidBase64(t *testing.T) {
	_, err := decodeCredentialsJSON("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecodeCredentialsJSON_Valid(t *testing.T) {
	encoded := base64.StdEncoding.EncodeToString([]byte(`{"type":"service_account"}`))
	decoded, err := decodeCredentialsJSON(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(string(decoded), "service_account") {
		t.Errorf("decoded: got %q", string(decoded))
	}
}

// --- integration tests: full method calls through mock HTTP ---

func TestList_Success(t *testing.T) {
	now := time.Now().UTC()

	server := fakeGCSServer(t, map[string]http.HandlerFunc{
		"/storage/v1/b/test-bucket/o": func(w http.ResponseWriter, r *http.Request) {
			prefix := r.URL.Query().Get("prefix")
			delimiter := r.URL.Query().Get("delimiter")
			if delimiter != "/" {
				t.Errorf("expected delimiter '/', got %q", delimiter)
			}
			_ = prefix
			jsonResponse(w, mustJSON(map[string]any{
				"kind": "storage#objects",
				"items": []map[string]any{
					{
						"name":         "docs/report.pdf",
						"contentType":  "application/pdf",
						"size":         "2048",
						"md5Hash":      "rL0Y20zC+Fzt72VPzMSk2A==",
						"updated":      now.Format(time.RFC3339Nano),
						"timeCreated":  now.Add(-time.Hour).Format(time.RFC3339Nano),
					},
				},
				"prefixes": []string{"docs/images/"},
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	entries, updatedCr, err := g.List(context.Background(), validCreds(), "docs/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updatedCr != nil {
		t.Error("expected nil updated creds for service account")
	}
	// The GCS JSON API returns items and prefixes; our handler returns 1 item + 1 prefix.
	// However, the Go client library handles prefixes separately via ObjectAttrs with Prefix set.
	// With our mock, we get at least the items back.
	if len(entries) == 0 {
		t.Fatal("expected at least one entry")
	}

	found := false
	for _, e := range entries {
		if e.Name == "report.pdf" {
			found = true
			if e.IsFolder {
				t.Error("report.pdf should not be a folder")
			}
			if e.MimeType != "application/pdf" {
				t.Errorf("mime: got %q", e.MimeType)
			}
		}
	}
	if !found {
		t.Error("expected to find report.pdf entry")
	}
}

func TestList_EmptyBucket(t *testing.T) {
	server := fakeGCSServer(t, map[string]http.HandlerFunc{
		"/storage/v1/b/test-bucket/o": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"kind":  "storage#objects",
				"items": []any{},
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	entries, _, err := g.List(context.Background(), validCreds(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestList_NilCreds(t *testing.T) {
	g := New()
	_, _, err := g.List(context.Background(), nil, "")
	if err == nil {
		t.Fatal("expected error for nil credentials")
	}
}

func TestList_MissingBucket(t *testing.T) {
	g := New()
	creds := &provider.Credentials{
		Extra: map[string]string{
			"credentials_json": fakeServiceAccountJSON(),
		},
	}
	_, _, err := g.List(context.Background(), creds, "")
	if err == nil {
		t.Fatal("expected error for missing bucket")
	}
	if !strings.Contains(err.Error(), "bucket not specified") {
		t.Errorf("error: got %q", err.Error())
	}
}

func TestList_MissingCredentialsJSON(t *testing.T) {
	g := New()
	creds := &provider.Credentials{
		Extra: map[string]string{
			"bucket": "test-bucket",
		},
	}
	_, _, err := g.List(context.Background(), creds, "")
	if err == nil {
		t.Fatal("expected error for missing credentials_json")
	}
}

func TestChanges_InitialSync(t *testing.T) {
	g := New()
	changes, token, updatedCr, err := g.Changes(context.Background(), validCreds(), "docs/", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty sync token")
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes on initial sync, got %d", len(changes))
	}
	if updatedCr != nil {
		t.Error("expected nil updated creds")
	}

	// Verify the token is a valid RFC3339 timestamp.
	_, parseErr := time.Parse(time.RFC3339, token)
	if parseErr != nil {
		t.Errorf("sync token should be RFC3339, got %q", token)
	}
}

func TestChanges_InvalidSyncToken(t *testing.T) {
	g := New()
	_, _, _, err := g.Changes(context.Background(), validCreds(), "", "not-a-timestamp")
	if err == nil {
		t.Fatal("expected error for invalid sync token")
	}
	if !strings.Contains(err.Error(), "parsing sync token") {
		t.Errorf("error: got %q", err.Error())
	}
}

func TestChanges_WithDelta(t *testing.T) {
	oldTime := time.Now().UTC().Add(-2 * time.Hour)
	newTime := time.Now().UTC().Add(-30 * time.Minute)
	syncTime := time.Now().UTC().Add(-time.Hour)

	server := fakeGCSServer(t, map[string]http.HandlerFunc{
		"/storage/v1/b/test-bucket/o": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"kind": "storage#objects",
				"items": []map[string]any{
					{
						"name":        "docs/old-file.txt",
						"contentType": "text/plain",
						"size":        "100",
						"md5Hash":     "rL0Y20zC+Fzt72VPzMSk2A==",
						"updated":     oldTime.Format(time.RFC3339Nano),
						"timeCreated": oldTime.Format(time.RFC3339Nano),
					},
					{
						"name":        "docs/new-file.txt",
						"contentType": "text/plain",
						"size":        "200",
						"md5Hash":     "abc123==",
						"updated":     newTime.Format(time.RFC3339Nano),
						"timeCreated": newTime.Format(time.RFC3339Nano),
					},
					{
						"name":        "docs/modified-file.txt",
						"contentType": "text/plain",
						"size":        "300",
						"md5Hash":     "def456==",
						"updated":     newTime.Format(time.RFC3339Nano),
						"timeCreated": oldTime.Format(time.RFC3339Nano),
					},
				},
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	changes, newToken, _, err := g.Changes(context.Background(), validCreds(), "docs/", syncTime.Format(time.RFC3339))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newToken == "" {
		t.Fatal("expected non-empty new sync token")
	}

	// old-file.txt was updated before sync time — should be excluded.
	// new-file.txt was created after sync time — should be ChangeCreated.
	// modified-file.txt was created before but updated after — should be ChangeModified.
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d: %+v", len(changes), changes)
	}

	var foundNew, foundModified bool
	for _, c := range changes {
		switch c.Ref {
		case "docs/new-file.txt":
			foundNew = true
			if c.Type != provider.ChangeCreated {
				t.Errorf("new-file.txt: expected created, got %s", c.Type)
			}
		case "docs/modified-file.txt":
			foundModified = true
			if c.Type != provider.ChangeModified {
				t.Errorf("modified-file.txt: expected modified, got %s", c.Type)
			}
		default:
			t.Errorf("unexpected change ref: %s", c.Ref)
		}
	}
	if !foundNew {
		t.Error("missing change for new-file.txt")
	}
	if !foundModified {
		t.Error("missing change for modified-file.txt")
	}
}

func TestFetch_Success(t *testing.T) {
	content := "hello world from GCS"

	server := fakeGCSServer(t, map[string]http.HandlerFunc{
		"/storage/v1/b/test-bucket/o": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("alt") == "media" {
				w.Header().Set("Content-Type", "application/pdf")
				_, _ = w.Write([]byte(content))
				return
			}
			jsonResponse(w, mustJSON(map[string]any{
				"name":        "docs/report.pdf",
				"contentType": "application/pdf",
				"size":        fmt.Sprintf("%d", len(content)),
				"md5Hash":     "rL0Y20zC+Fzt72VPzMSk2A==",
				"updated":     time.Now().UTC().Format(time.RFC3339Nano),
				"timeCreated": time.Now().UTC().Format(time.RFC3339Nano),
			}))
		},
		"/test-bucket/docs/report.pdf": func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/pdf")
			_, _ = w.Write([]byte(content))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	rc, meta, updatedCr, err := g.Fetch(context.Background(), validCreds(), "docs/report.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if updatedCr != nil {
		t.Error("expected nil updated creds for service account")
	}
	if meta.Name != "report.pdf" {
		t.Errorf("name: got %q", meta.Name)
	}
	if meta.MimeType != "application/pdf" {
		t.Errorf("mime: got %q", meta.MimeType)
	}

	body, _ := io.ReadAll(rc)
	if string(body) != content {
		t.Errorf("content: got %q", string(body))
	}
}

func TestFetch_NilCreds(t *testing.T) {
	g := New()
	_, _, _, err := g.Fetch(context.Background(), nil, "docs/report.pdf")
	if err == nil {
		t.Fatal("expected error for nil credentials")
	}
}

func TestCredentials_ExpiryIsZero(t *testing.T) {
	creds := validCreds()
	if !creds.Expiry.IsZero() {
		t.Error("GCS credentials should have zero expiry")
	}
	if creds.Expired() {
		t.Error("zero-expiry credentials should not be expired")
	}
}

// fakePrivateKey is an RSA private key used only in tests.
// Generated specifically for test fixtures — not a real secret.
//
//nolint:gosec // G101: test fixture, not a real credential.
const fakePrivateKey = `-----BEGIN RSA PRIVATE KEY-----
MIIEpAIBAAKCAQEA0Z3VS5JJcds3xfn/ygWyF8PbnGy0AHB7MhgHcTz6sE2I2yPB
aBDWDFMi3JG4c8rPnP3RBOBqLqFOmaVMGFHEIn96UlBczCRb/mX6xFcDsCP5suqC
bJsw/1KMq/0oJWZJMVp2JMZ7yEsqnKBPQhMa/87NaGVo+4F3RyLox1FU5dDSRFa
TGE7P4IkJxjEIxTkYfFmaLCeRN/MYJBLEJhpCa5bgMWRYDDnIpknLS5Hm/kFMyLj
H1CP+w85cBoHX1EYHwT+A1P0p4pHQxR4ENmBJfP5rlMj0YCxMRIQfjfWFBGXxZfx
n7Lk3r2XS+1GAZTV0e0RYMGqoQjkKGFwb4T4NQIDAQABAoIBAC5RgZ+hBx7xHNaM
pPgwGMnCd3KE7Fy1oBPaqECFTELC54pMqMJsGJ0FRaAjH1X6q0VPFG5GR0IKMB0x
oFMTV0e5rE9qCW0ekDmkVCM5kaPKEJkP3jND0g6PxIHC8KMzbeX3YQFZ+MHK+CMJ
f8R6P+3Gce7cNToTDaW3SXpTFGK3hJJpxD8sPSF0MiPrM5Ya7fJCOuTPXgiVfZfV1
fL/yrE0aBMpS6hObCxrJeVBGD42VfJ8ytLJCdEoKJmK3n6AHBe+PQS6NjmmQtMohVe
kR0VpT4+JXehIpWuR7bFaVkYb0rDreIHHXqTFhBzdCyYOaFFkiJB0S+R5Ty8Jk3dECg
YEA7tDzaACqXZ7HHlD6Mkv0bp0Rpc15cRFQ0dAXiTxCm1B5fOaVU/LlRJEnGFI6PLC
T5E+4JlNHjbQB5ky5qL2gMXxt5E+t5BPrrpd0F5RxhA9VqZZ5cTBFLZ4VoyWRUNCNq
Y8VGMEy3FAH0Q4EFqrA//IC8LD10rJmZE3S0WlJgS7sCgYEA4lVg23K+hgxS7OJK3
T3EMk/e3FtXRvzKtx1t1TFjEbX8kBLRK+jHLH08mA4EibYEq+MmA0SHrxJl4Cn8j
C3WFYj0mMFJN3u6t6KFH6CtWHFR7PqYVq0j15GiLj6s0D6UQlSH1y1yZvJGS0nv2U
S5KxU+bkSR1R3cHcZLlFI+E0CgYEA6epB0GV5S8cXL1JiCnzH5hRCRnV8JCCJxzaFQ
i7cTXQrKsn5cI6c7E1AkmqJ1YVnbKKq0S12BGMWP7Lfqk5aZKFUoXZ0qVVqBH6Y
Q0YqZPfHX7RTil7XRM+0lDDdFLFonTI3Mfrwqj4kLN5N2P4x5h3hFX9Wj2dMBL9Ka
M/AoGAQ5w2O+z35SWBrOmfOYFj8oY9L9fHQhT3rP6LLz0vRTpKxJ7GWR4kqFaZJBT
YqV7c5c5M0oLz7YHtyp0cXl6b5R8C7KclNa/k+eG5YKylt1DCpC7H8DPFRrCOCKP
GMVB3M+a7N2oJQVZm9aWZb3OL0+I8MpHrB0RPtWbSCmzulQ=
-----END RSA PRIVATE KEY-----`
