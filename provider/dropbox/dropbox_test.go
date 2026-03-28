package dropbox

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*Dropbox)(nil)

// --- test helpers ---

// fakeServer returns an httptest.Server that routes by path prefix.
func fakeServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for prefix, handler := range handlers {
			if strings.HasPrefix(r.URL.Path, prefix) {
				handler(w, r)
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

// newTestProvider creates a Dropbox pointed at a test server.
func newTestProvider(t *testing.T, server *httptest.Server) *Dropbox {
	t.Helper()
	return &Dropbox{
		oauth: &oauth2.Config{
			ClientID:     "test-client",
			ClientSecret: "test-secret",
			Endpoint: oauth2.Endpoint{
				TokenURL: server.URL + "/oauth2/token",
			},
		},
		apiEndpoint:     server.URL,
		contentEndpoint: server.URL,
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
	d := New(Config{ClientID: "id", ClientSecret: "secret"})
	if d.Type() != "dropbox" {
		t.Errorf("type: got %q, want %q", d.Type(), "dropbox")
	}
}

func TestAuthURL(t *testing.T) {
	d := New(Config{ClientID: "client-id", ClientSecret: "secret"})
	url, err := d.AuthURL(context.Background(), "https://app.example.com/callback", "csrf-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "client_id=client-id") {
		t.Errorf("URL should contain client_id, got %s", url)
	}
	if !strings.Contains(url, "state=csrf-token") {
		t.Errorf("URL should contain state, got %s", url)
	}
	if !strings.Contains(url, "dropbox.com") {
		t.Errorf("URL should point to dropbox.com, got %s", url)
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{"", ""},
		{"/", ""},
		{"/Documents", "/Documents"},
		{"Documents", "/Documents"},
		{"/Documents/Sub", "/Documents/Sub"},
	}
	for _, tt := range tests {
		if got := normalizePath(tt.input); got != tt.expected {
			t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.expected)
		}
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

func TestMetadataToEntry_File(t *testing.T) {
	m := &fileMetadata{
		Tag: "file", Name: "report.pdf", PathDisplay: "/Documents/report.pdf",
		Size: 1024, ContentHash: "abc123", ServerModified: "2025-06-01T12:00:00Z",
	}
	entry := metadataToEntry(m)
	if entry.Ref != "/Documents/report.pdf" || entry.Name != "report.pdf" || entry.IsFolder {
		t.Errorf("unexpected entry: %+v", entry)
	}
	if entry.ContentHash != "abc123" {
		t.Errorf("hash: got %q", entry.ContentHash)
	}
	if entry.MimeType != "application/pdf" {
		t.Errorf("mime: got %q", entry.MimeType)
	}
}

func TestMetadataToEntry_Folder(t *testing.T) {
	m := &fileMetadata{Tag: "folder", Name: "Documents", PathDisplay: "/Documents"}
	entry := metadataToEntry(m)
	if !entry.IsFolder {
		t.Error("expected folder")
	}
	if entry.MimeType != "" {
		t.Errorf("folder should have empty mime, got %q", entry.MimeType)
	}
}

func TestMetadataToChange_Deleted(t *testing.T) {
	m := &fileMetadata{Tag: "deleted", Name: "old.pdf", PathDisplay: "/old.pdf"}
	change := metadataToChange(m)
	if change == nil || change.Type != provider.ChangeDeleted {
		t.Errorf("expected deleted change, got %v", change)
	}
	if change.Ref != "/old.pdf" {
		t.Errorf("ref: got %q", change.Ref)
	}
}

func TestMetadataToChange_Modified(t *testing.T) {
	m := &fileMetadata{
		Tag: "file", Name: "doc.pdf", PathDisplay: "/doc.pdf",
		Size: 512, ServerModified: "2025-06-01T12:00:00Z",
	}
	change := metadataToChange(m)
	if change == nil || change.Type != provider.ChangeModified {
		t.Errorf("expected modified change, got %v", change)
	}
	if change.Entry == nil {
		t.Fatal("expected entry in change")
	}
}

func TestGuessMIME(t *testing.T) {
	tests := []struct{ input, expected string }{
		{"report.pdf", "application/pdf"},
		{"data.csv", "text/csv"},
		{"photo.PNG", "image/png"},
		{"image.jpg", "image/jpeg"},
		{"image.JPEG", "image/jpeg"},
		{"config.json", "application/json"},
		{"readme.txt", "text/plain"},
		{"spreadsheet.xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"unknown.xyz", "application/octet-stream"},
	}
	for _, tt := range tests {
		if got := guessMIME(tt.input); got != tt.expected {
			t.Errorf("guessMIME(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

// --- integration tests: full method calls through mock HTTP ---

func TestExchange_Success(t *testing.T) {
	server := httptest.NewServer(tokenHandler())
	defer server.Close()

	d := &Dropbox{
		oauth: &oauth2.Config{
			ClientID: "id", ClientSecret: "secret",
			Endpoint:    oauth2.Endpoint{TokenURL: server.URL},
			RedirectURL: "https://app.example.com/callback",
		},
	}

	creds, err := d.Exchange(context.Background(), "auth-code", "https://app.example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessToken != "test-access-token" {
		t.Errorf("access token: got %q", creds.AccessToken)
	}
}

func TestList_Success(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
		"/2/files/list_folder": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(listFolderResponse{
				Entries: []fileMetadata{
					{Tag: "file", Name: "doc.pdf", PathDisplay: "/Documents/doc.pdf", Size: 2048, ServerModified: "2025-06-01T12:00:00Z"},
					{Tag: "folder", Name: "Photos", PathDisplay: "/Documents/Photos"},
				},
				Cursor:  "cursor-1",
				HasMore: false,
			}))
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	entries, updatedCr, err := d.List(context.Background(), validCreds(), "/Documents")
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
		"/oauth2/token": tokenHandler(),
		"/2/files/list_folder": func(w http.ResponseWriter, r *http.Request) {
			var body listFolderRequest
			if decErr := json.NewDecoder(r.Body).Decode(&body); decErr != nil {
				t.Errorf("decoding body: %v", decErr)
			}
			if body.Path != "" {
				t.Errorf("expected empty path for root, got %q", body.Path)
			}
			jsonResponse(w, mustJSON(listFolderResponse{
				Entries: []fileMetadata{},
				Cursor:  "cursor-root",
				HasMore: false,
			}))
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	entries, _, err := d.List(context.Background(), validCreds(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestList_Pagination(t *testing.T) {
	callCount := 0
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
		"/2/files/list_folder": func(w http.ResponseWriter, _ *http.Request) {
			callCount++
			if callCount == 1 {
				jsonResponse(w, mustJSON(listFolderResponse{
					Entries: []fileMetadata{
						{Tag: "file", Name: "page1.pdf", PathDisplay: "/page1.pdf"},
					},
					Cursor:  "cursor-page2",
					HasMore: true,
				}))
			} else {
				jsonResponse(w, mustJSON(listFolderResponse{
					Entries: []fileMetadata{
						{Tag: "file", Name: "page2.pdf", PathDisplay: "/page2.pdf"},
					},
					Cursor:  "cursor-done",
					HasMore: false,
				}))
			}
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	entries, _, err := d.List(context.Background(), validCreds(), "/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries across pages, got %d", len(entries))
	}
}

func TestChanges_InitialSync(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
		"/2/files/list_folder/get_latest_cursor": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]string{"cursor": "initial-cursor"}))
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	changes, cursor, _, err := d.Changes(context.Background(), validCreds(), "/Documents", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cursor != "initial-cursor" {
		t.Errorf("cursor: got %q", cursor)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes on initial sync, got %d", len(changes))
	}
}

func TestChanges_WithDelta(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
		"/2/files/list_folder/continue": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(listFolderResponse{
				Entries: []fileMetadata{
					{Tag: "file", Name: "updated.pdf", PathDisplay: "/Documents/updated.pdf", Size: 512, ServerModified: "2025-06-02T12:00:00Z"},
					{Tag: "deleted", Name: "removed.pdf", PathDisplay: "/Documents/removed.pdf"},
				},
				Cursor:  "cursor-2",
				HasMore: false,
			}))
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	changes, newCursor, _, err := d.Changes(context.Background(), validCreds(), "/Documents", "cursor-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newCursor != "cursor-2" {
		t.Errorf("new cursor: got %q", newCursor)
	}
	if len(changes) != 2 {
		t.Fatalf("expected 2 changes, got %d", len(changes))
	}
	if changes[0].Ref != "/Documents/updated.pdf" || changes[0].Type != provider.ChangeModified {
		t.Errorf("change 0: %+v", changes[0])
	}
	if changes[1].Ref != "/Documents/removed.pdf" || changes[1].Type != provider.ChangeDeleted {
		t.Errorf("change 1: %+v", changes[1])
	}
}

func TestFetch_Success(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
		"/2/files/download": func(w http.ResponseWriter, r *http.Request) {
			apiArg := r.Header.Get("Dropbox-API-Arg")
			if apiArg == "" {
				t.Error("missing Dropbox-API-Arg header")
			}
			var arg map[string]string
			if err := json.Unmarshal([]byte(apiArg), &arg); err != nil {
				t.Errorf("invalid Dropbox-API-Arg: %v", err)
			}
			if arg["path"] != "/Documents/report.pdf" {
				t.Errorf("path: got %q", arg["path"])
			}

			meta := fileMetadata{
				Name: "report.pdf", Size: 9, ContentHash: "hash123",
			}
			w.Header().Set("Dropbox-API-Result", string(mustJSON(meta)))
			_, _ = w.Write([]byte("pdf bytes"))
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	rc, meta, _, err := d.Fetch(context.Background(), validCreds(), "/Documents/report.pdf")
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
	if meta.MimeType != "application/pdf" {
		t.Errorf("mime: got %q", meta.MimeType)
	}

	body, _ := io.ReadAll(rc)
	if string(body) != "pdf bytes" {
		t.Errorf("content: got %q", string(body))
	}
}

func TestFetch_EmptyRef(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
	})
	defer server.Close()

	d := newTestProvider(t, server)
	_, _, _, err := d.Fetch(context.Background(), validCreds(), "")
	if err == nil {
		t.Fatal("expected error for empty ref")
	}
	if !strings.Contains(err.Error(), "empty file reference") {
		t.Errorf("error should mention empty ref, got %q", err.Error())
	}
}

func TestFetch_HTTPError(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/oauth2/token": tokenHandler(),
		"/2/files/download": func(w http.ResponseWriter, _ *http.Request) {
			http.Error(w, "not found", http.StatusNotFound)
		},
	})
	defer server.Close()

	d := newTestProvider(t, server)
	_, _, _, err := d.Fetch(context.Background(), validCreds(), "/missing.pdf")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !strings.Contains(err.Error(), "status 404") {
		t.Errorf("error should mention 404, got %q", err.Error())
	}
}
