package onedrive

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

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*OneDrive)(nil)

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

// newTestProvider creates a OneDrive pointed at a test server.
func newTestProvider(t *testing.T, server *httptest.Server) *OneDrive {
	t.Helper()
	return &OneDrive{
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
	o := New(Config{ClientID: "id", ClientSecret: "secret"})
	if o.Type() != "onedrive" {
		t.Errorf("type: got %q, want %q", o.Type(), "onedrive")
	}
}

func TestAuthURL(t *testing.T) {
	o := New(Config{ClientID: "client-id", ClientSecret: "secret"})
	u, err := o.AuthURL(context.Background(), "https://app.example.com/callback", "csrf-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(u, "client_id=client-id") {
		t.Errorf("URL should contain client_id, got %s", u)
	}
	if !strings.Contains(u, "state=csrf-token") {
		t.Errorf("URL should contain state, got %s", u)
	}
	if !strings.Contains(u, "access_type=offline") {
		t.Errorf("URL should request offline access, got %s", u)
	}
}

func TestDefaultScopes(t *testing.T) {
	scopes := DefaultScopes()
	if len(scopes) != 2 {
		t.Fatalf("expected 2 scopes, got %d", len(scopes))
	}
	if scopes[0] != "Files.Read.All" {
		t.Errorf("expected Files.Read.All scope, got %q", scopes[0])
	}
	if scopes[1] != "offline_access" {
		t.Errorf("expected offline_access scope, got %q", scopes[1])
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

func TestItemToEntry_File(t *testing.T) {
	item := &graphItem{
		ID: "item-1", Name: "report.pdf", Size: 1024,
		LastModifiedAt: time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC),
		File: &graphFile{
			MIMEType: "application/pdf",
			Hashes:   &fileHashes{SHA256Hash: "abc123"},
		},
	}
	entry := itemToEntry(item)
	if entry.Ref != "item-1" || entry.Name != "report.pdf" || entry.IsFolder {
		t.Errorf("unexpected entry: %+v", entry)
	}
	if entry.ContentHash != "abc123" {
		t.Errorf("hash: got %q", entry.ContentHash)
	}
}

func TestItemToEntry_Folder(t *testing.T) {
	item := &graphItem{
		ID: "folder-1", Name: "Documents",
		Folder: &graphFolder{ChildCount: 5},
	}
	if !itemToEntry(item).IsFolder {
		t.Error("expected folder")
	}
}

func TestHashFromItem_Preference(t *testing.T) {
	// SHA256 preferred over SHA1.
	item := &graphItem{
		File: &graphFile{Hashes: &fileHashes{SHA1Hash: "sha1", SHA256Hash: "sha256"}},
	}
	if got := hashFromItem(item); got != "sha256" {
		t.Errorf("expected sha256, got %q", got)
	}

	// SHA1 when no SHA256.
	item.File.Hashes.SHA256Hash = ""
	if got := hashFromItem(item); got != "sha1" {
		t.Errorf("expected sha1, got %q", got)
	}

	// QuickXOR as fallback.
	item.File.Hashes.SHA1Hash = ""
	item.File.Hashes.QuickXOR = "qxor"
	if got := hashFromItem(item); got != "qxor" {
		t.Errorf("expected qxor, got %q", got)
	}

	// No hashes.
	if got := hashFromItem(&graphItem{}); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestMimeFromItem(t *testing.T) {
	if got := mimeFromItem(&graphItem{File: &graphFile{MIMEType: "text/plain"}}); got != "text/plain" {
		t.Errorf("expected text/plain, got %q", got)
	}
	if got := mimeFromItem(&graphItem{Folder: &graphFolder{}}); got != "application/vnd.ms-folder" {
		t.Errorf("expected folder mime, got %q", got)
	}
	if got := mimeFromItem(&graphItem{}); got != "" {
		t.Errorf("expected empty, got %q", got)
	}
}

func TestItemToChange_Deleted(t *testing.T) {
	item := &graphItem{ID: "item-1", Deleted: &struct{}{}}
	change := itemToChange(item)
	if change.Type != provider.ChangeDeleted {
		t.Errorf("expected deleted, got %q", change.Type)
	}
	if change.Entry != nil {
		t.Error("expected nil entry for deletion")
	}
}

func TestItemToChange_Modified(t *testing.T) {
	item := &graphItem{
		ID: "item-1", Name: "doc.pdf",
		File: &graphFile{MIMEType: "application/pdf"},
	}
	change := itemToChange(item)
	if change.Type != provider.ChangeModified {
		t.Errorf("expected modified, got %q", change.Type)
	}
	if change.Entry == nil || change.Entry.Name != "doc.pdf" {
		t.Errorf("unexpected entry: %+v", change.Entry)
	}
}

// --- integration tests: full method calls through mock HTTP ---

func TestExchange_Success(t *testing.T) {
	server := httptest.NewServer(tokenHandler())
	defer server.Close()

	o := &OneDrive{
		oauth: &oauth2.Config{
			ClientID: "id", ClientSecret: "secret",
			Endpoint:    oauth2.Endpoint{TokenURL: server.URL},
			RedirectURL: "https://app.example.com/callback",
		},
	}

	creds, err := o.Exchange(context.Background(), "auth-code", "https://app.example.com/callback")
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
		"/me/drive/items/root/children": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"value": []map[string]any{
					{
						"id": "f1", "name": "doc.pdf", "size": 2048,
						"lastModifiedDateTime": "2025-06-01T12:00:00Z",
						"file": map[string]any{
							"mimeType": "application/pdf",
							"hashes":   map[string]any{"sha256Hash": "deadbeef"},
						},
					},
					{
						"id": "f2", "name": "Photos",
						"lastModifiedDateTime": "2025-06-02T12:00:00Z",
						"folder":               map[string]any{"childCount": 3},
					},
				},
			}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)
	entries, updatedCr, err := o.List(context.Background(), validCreds(), "root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].Name != "doc.pdf" {
		t.Errorf("entry 0 name: got %q", entries[0].Name)
	}
	if entries[0].ContentHash != "deadbeef" {
		t.Errorf("entry 0 hash: got %q", entries[0].ContentHash)
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
		"/me/drive/items/root/children": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{"value": []any{}}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)
	entries, _, err := o.List(context.Background(), validCreds(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestList_InvalidItemID(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
	})
	defer server.Close()

	o := newTestProvider(t, server)
	_, _, err := o.List(context.Background(), validCreds(), "item'--injection")
	if err == nil {
		t.Fatal("expected error for invalid item ID")
	}
	if !strings.Contains(err.Error(), "invalid item ID") {
		t.Errorf("error should mention invalid item ID, got %q", err.Error())
	}
}

func TestList_Pagination(t *testing.T) {
	callCount := 0
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/me/drive/items/root/children": func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if callCount == 1 {
				jsonResponse(w, mustJSON(map[string]any{
					"value": []map[string]any{
						{"id": "f1", "name": "page1.pdf", "file": map[string]any{"mimeType": "application/pdf"}},
					},
					"@odata.nextLink": "http://" + r.Host + "/me/drive/items/root/children?page=2",
				}))
				return
			}
			jsonResponse(w, mustJSON(map[string]any{
				"value": []map[string]any{
					{"id": "f2", "name": "page2.pdf", "file": map[string]any{"mimeType": "application/pdf"}},
				},
			}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)
	entries, _, err := o.List(context.Background(), validCreds(), "root")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries across pages, got %d", len(entries))
	}
}

func TestChanges_InitialSync(t *testing.T) {
	wantDelta := "https://graph.microsoft.com/delta?token=start-1" //nolint:gosec // test URL, not a credential
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/me/drive/root/delta": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"value":            []any{},
				"@odata.deltaLink": wantDelta,
			}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)
	changes, token, _, err := o.Changes(context.Background(), validCreds(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != wantDelta {
		t.Errorf("sync token: got %q", token)
	}
	if len(changes) != 0 {
		t.Errorf("expected 0 changes on initial sync, got %d", len(changes))
	}
}

func TestChanges_WithDelta(t *testing.T) {
	wantDelta := "https://graph.microsoft.com/delta?token=token-2" //nolint:gosec // test URL, not a credential
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/me/drive/root/delta": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"@odata.deltaLink": wantDelta,
				"value": []map[string]any{
					{
						"id": "f1", "name": "updated.pdf",
						"file": map[string]any{"mimeType": "application/pdf"},
					},
					{
						"id": "f2", "name": "deleted-item",
						"deleted": map[string]any{},
					},
				},
			}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)

	// Use the server URL as deltaLink since the test server handles the path.
	deltaLink := server.URL + "/me/drive/root/delta"
	changes, newToken, _, err := o.Changes(context.Background(), validCreds(), "", deltaLink)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if newToken != wantDelta {
		t.Errorf("new token: got %q", newToken)
	}
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

func TestChanges_FolderScoped(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/me/drive/root:/Documents:/delta": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"value":              []any{},
				"@odata.deltaLink": "delta-link-scoped",
			}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)
	_, token, _, err := o.Changes(context.Background(), validCreds(), "Documents", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "delta-link-scoped" {
		t.Errorf("expected scoped delta link, got %q", token)
	}
}

func TestFetch_RegularFile(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
		"/me/drive/items/": func(w http.ResponseWriter, r *http.Request) {
			if strings.HasSuffix(r.URL.Path, "/content") {
				_, _ = w.Write([]byte("pdf bytes"))
				return
			}
			jsonResponse(w, mustJSON(map[string]any{
				"id": "file-1", "name": "report.pdf", "size": 9,
				"file": map[string]any{
					"mimeType": "application/pdf",
					"hashes":   map[string]any{"sha1Hash": "hash123"},
				},
			}))
		},
	})
	defer server.Close()

	o := newTestProvider(t, server)
	rc, meta, _, err := o.Fetch(context.Background(), validCreds(), "file-1")
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

func TestFetch_InvalidItemID(t *testing.T) {
	server := fakeServer(t, map[string]http.HandlerFunc{
		"/token": tokenHandler(),
	})
	defer server.Close()

	o := newTestProvider(t, server)
	_, _, _, err := o.Fetch(context.Background(), validCreds(), "bad/id")
	if err == nil {
		t.Fatal("expected error for invalid item ID")
	}
	if !strings.Contains(err.Error(), "invalid item ID") {
		t.Errorf("error should mention invalid item ID, got %q", err.Error())
	}
}
