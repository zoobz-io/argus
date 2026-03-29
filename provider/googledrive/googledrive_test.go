package googledrive

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*GoogleDrive)(nil)

// --- test helpers ---

// fakeServer returns an httptest.Server that routes by path prefix.
// Routes are matched longest-prefix-first to avoid ambiguity.
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

// tokenHandler returns an HTTP handler that serves valid OAuth2 token responses.
func tokenHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, mustJSON(map[string]any{
			"access_token": "test-access-token",
			"token_type":   "Bearer",
			"expires_in":   3600,
		}))
	}
}

// newTestProvider creates a GoogleDrive pointed at a test server.
func newTestProvider(t *testing.T, server *httptest.Server) *GoogleDrive {
	t.Helper()
	return &GoogleDrive{
		oauth: &oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			Endpoint: oauth2.Endpoint{
				TokenURL: server.URL + "/token",
			},
		},
		endpoint: server.URL,
	}
}

// validCreds returns non-expired test credentials.
func validCreds() *provider.Credentials {
	return &provider.Credentials{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		Expiry:       time.Now().Add(time.Hour),
	}
}

// --- unit tests: helpers ---

func TestType(t *testing.T) {
	g := New(Config{ClientID: "id", ClientSecret: "secret"})
	if g.Type() != "google_drive" {
		t.Errorf("type: got %q, want %q", g.Type(), "google_drive")
	}
}

func TestAuthURL(t *testing.T) {
	g := New(Config{ClientID: "client-id", ClientSecret: "secret"})
	url, err := g.AuthURL(context.Background(), "https://app.example.com/callback", "csrf-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "client_id=client-id") {
		t.Errorf("URL should contain client_id, got %s", url)
	}
	if !strings.Contains(url, "state=csrf-token") {
		t.Errorf("URL should contain state, got %s", url)
	}
	if !strings.Contains(url, "access_type=offline") {
		t.Errorf("URL should request offline access, got %s", url)
	}
}

func TestDefaultScopes(t *testing.T) {
	scopes := DefaultScopes()
	if len(scopes) == 0 {
		t.Fatal("expected at least one scope")
	}
	if scopes[0] != drive.DriveReadonlyScope {
		t.Errorf("expected readonly scope, got %q", scopes[0])
	}
}

func TestTokenToCreds(t *testing.T) {
	expiry := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	token := &oauth2.Token{
		AccessToken: "access", RefreshToken: "refresh",
		TokenType: "Bearer", Expiry: expiry,
	}
	creds := tokenToCreds(token)
	if creds.AccessToken != "access" {
		t.Errorf("AccessToken: got %q", creds.AccessToken)
	}
	if creds.RefreshToken != "refresh" {
		t.Errorf("RefreshToken: got %q", creds.RefreshToken)
	}
}

func TestCredsToToken(t *testing.T) {
	creds := &provider.Credentials{
		AccessToken: "access", RefreshToken: "refresh",
		TokenType: "Bearer", Expiry: time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	token := credsToToken(creds)
	if token.AccessToken != "access" {
		t.Errorf("AccessToken: got %q", token.AccessToken)
	}
}

func TestUpdatedCreds_NoChange(t *testing.T) {
	original := &provider.Credentials{AccessToken: "same"}
	current := &oauth2.Token{AccessToken: "same"}
	if updatedCreds(original, current) != nil {
		t.Error("expected nil when token unchanged")
	}
}

func TestUpdatedCreds_Changed(t *testing.T) {
	original := &provider.Credentials{AccessToken: "old"}
	current := &oauth2.Token{AccessToken: "new", RefreshToken: "refresh"}
	if updatedCreds(original, current) == nil {
		t.Fatal("expected updated creds")
	}
}

func TestUpdatedCreds_NilToken(t *testing.T) {
	original := &provider.Credentials{AccessToken: "old"}
	if updatedCreds(original, nil) != nil {
		t.Error("expected nil for nil token")
	}
}

func TestFileToEntry(t *testing.T) {
	f := &drive.File{
		Id: "file-1", Name: "report.pdf", MimeType: "application/pdf",
		Size: 1024, Md5Checksum: "abc123", ModifiedTime: "2025-06-01T12:00:00Z",
	}
	entry := fileToEntry(f)
	if entry.Ref != "file-1" || entry.Name != "report.pdf" || entry.IsFolder {
		t.Errorf("unexpected entry: %+v", entry)
	}
}

func TestFileToEntry_Folder(t *testing.T) {
	f := &drive.File{Id: "folder-1", Name: "Documents", MimeType: "application/vnd.google-apps.folder"}
	if !fileToEntry(f).IsFolder {
		t.Error("expected folder")
	}
}

func TestIsGoogleNative(t *testing.T) {
	if !isGoogleNative("application/vnd.google-apps.document") {
		t.Error("expected native")
	}
	if isGoogleNative("application/pdf") {
		t.Error("expected not native")
	}
}

func TestGoogleExportMIME(t *testing.T) {
	tests := []struct{ input, expected string }{
		{"application/vnd.google-apps.document", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"application/vnd.google-apps.spreadsheet", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"application/vnd.google-apps.presentation", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{"application/vnd.google-apps.drawing", "application/pdf"},
	}
	for _, tt := range tests {
		if got := googleExportMIME(tt.input); got != tt.expected {
			t.Errorf("googleExportMIME(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestChangeToProviderChange_Removed(t *testing.T) {
	c := &drive.Change{FileId: "file-1", Removed: true}
	change := changeToProviderChange(c, "")
	if change == nil || change.Type != provider.ChangeDeleted {
		t.Errorf("expected deleted change, got %v", change)
	}
}

func TestChangeToProviderChange_NilFile(t *testing.T) {
	if changeToProviderChange(&drive.Change{FileId: "f1"}, "") != nil {
		t.Error("expected nil for change with no file and not removed")
	}
}

func TestChangeToProviderChange_FilteredByPath(t *testing.T) {
	c := &drive.Change{FileId: "f1", File: &drive.File{Id: "f1", Parents: []string{"other"}}}
	if changeToProviderChange(c, "my-folder") != nil {
		t.Error("expected nil for file in different folder")
	}
}

func TestChangeToProviderChange_MatchesPath(t *testing.T) {
	c := &drive.Change{FileId: "f1", File: &drive.File{Id: "f1", Name: "doc.pdf", MimeType: "application/pdf", Parents: []string{"my-folder"}}}
	change := changeToProviderChange(c, "my-folder")
	if change == nil || change.Type != provider.ChangeModified {
		t.Errorf("expected modified change, got %v", change)
	}
}

// --- integration tests: full method calls through mock HTTP ---

func TestExchange_Success(t *testing.T) {
	server := httptest.NewServer(tokenHandler())
	defer server.Close()

	g := &GoogleDrive{
		oauth: &oauth2.Config{
			ClientID: "id", ClientSecret: "secret",
			Endpoint:    oauth2.Endpoint{TokenURL: server.URL},
			RedirectURL: "https://app.example.com/callback",
		},
	}

	creds, err := g.Exchange(context.Background(), "auth-code", "https://app.example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessToken != "test-access-token" {
		t.Errorf("access token: got %q", creds.AccessToken)
	}
}

func TestList_Success(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/files": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"files": []map[string]any{
					{"id": "f1", "name": "doc.pdf", "mimeType": "application/pdf", "size": "2048", "modifiedTime": "2025-06-01T12:00:00Z"},
					{"id": "f2", "name": "Photos", "mimeType": "application/vnd.google-apps.folder", "modifiedTime": "2025-06-02T12:00:00Z"},
				},
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	entries, updatedCr, err := g.List(context.Background(), validCreds(), "root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "doc.pdf" {
		t.Errorf("entry 0 name: got %q", entries[0].Name)
	}
	if !entries[1].IsFolder {
		t.Error("entry 1 should be folder")
	}
	// Token unchanged — should be nil.
	if updatedCr != nil {
		t.Error("expected nil updated creds when token unchanged")
	}
}

func TestList_EmptyPath(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/files": func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("q")
			if !strings.Contains(q, "'root' in parents") {
				t.Errorf("expected root query, got %q", q)
			}
			jsonResponse(w, mustJSON(map[string]any{"files": []any{}}))
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

func TestList_InvalidFolderID(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
	})
	defer server.Close()

	g := newTestProvider(t, server)
	_, _, err := g.List(context.Background(), validCreds(), "folder'--injection")
	if err == nil {
		t.Fatal("expected error for invalid folder ID")
	}
	if !strings.Contains(err.Error(), "invalid folder ID") {
		t.Errorf("error should mention invalid folder ID, got %q", err.Error())
	}
}

func TestChanges_InitialSync(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token":                  tokenHandler(),
		"/changes/startPageToken": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{"startPageToken": "start-1"}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	changes, token, _, err := g.Changes(context.Background(), validCreds(), "folder-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "start-1" {
		t.Errorf("sync token: got %q", token)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes on initial sync, got %d", len(changes))
	}
}

func TestChanges_WithDelta(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/changes": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"newStartPageToken": "token-2",
				"changes": []map[string]any{
					{"fileId": "f1", "removed": false, "file": map[string]any{
						"id": "f1", "name": "updated.pdf", "mimeType": "application/pdf", "parents": []string{"watched"},
					}},
					{"fileId": "f2", "removed": true},
					{"fileId": "f3", "removed": false, "file": map[string]any{
						"id": "f3", "name": "other.pdf", "mimeType": "application/pdf", "parents": []string{"different"},
					}},
				},
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	changes, newToken, _, err := g.Changes(context.Background(), validCreds(), "watched", "token-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newToken != "token-2" {
		t.Errorf("new token: got %q", newToken)
	}
	// f1 matches folder, f2 deleted (always included), f3 wrong folder (filtered).
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].Ref != "f1" || changes[0].Type != provider.ChangeModified {
		t.Errorf("change 0: %+v", changes[0])
	}
	if changes[1].Ref != "f2" || changes[1].Type != provider.ChangeDeleted {
		t.Errorf("change 1: %+v", changes[1])
	}
}

func TestFetch_RegularFile(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/files/file-1": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("alt") == "media" {
				_, _ = w.Write([]byte("pdf bytes"))
				return
			}
			jsonResponse(w, mustJSON(map[string]any{
				"id": "file-1", "name": "report.pdf", "mimeType": "application/pdf",
				"size": "9", "md5Checksum": "hash123",
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	rc, meta, _, err := g.Fetch(context.Background(), validCreds(), "file-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if meta.Name != "report.pdf" {
		t.Errorf("name: got %q", meta.Name)
	}
	if meta.ContentHash != "hash123" {
		t.Errorf("hash: got %q", meta.ContentHash)
	}

	body, _ := io.ReadAll(rc)
	if string(body) != "pdf bytes" {
		t.Errorf("content: got %q", string(body))
	}
}

func TestFetch_GoogleNativeExport(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/files/doc-1/export": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("exported docx"))
		},
		"/files/doc-1": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"id": "doc-1", "name": "My Doc",
				"mimeType": "application/vnd.google-apps.document", "size": "0",
			}))
		},
	})
	defer server.Close()

	g := newTestProvider(t, server)
	rc, meta, _, err := g.Fetch(context.Background(), validCreds(), "doc-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer func() { _ = rc.Close() }()

	if meta.MimeType != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Errorf("mime: got %q", meta.MimeType)
	}

	body, _ := io.ReadAll(rc)
	if string(body) != "exported docx" {
		t.Errorf("content: got %q", string(body))
	}
}
