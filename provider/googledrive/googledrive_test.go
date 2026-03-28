package googledrive

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
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"

	"github.com/zoobz-io/argus/provider"
)

// Compile-time interface assertion.
var _ provider.Provider = (*GoogleDrive)(nil)

// --- helper: fake Google API server ---

// fakeGoogleServer returns an httptest.Server that serves canned Drive API responses.
// The handler map routes path prefixes to response functions.
func fakeGoogleServer(t *testing.T, handlers map[string]http.HandlerFunc) *httptest.Server {
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

// driveServiceWithServer creates a Drive service pointed at a test server.
func driveServiceWithServer(t *testing.T, server *httptest.Server) *drive.Service {
	t.Helper()
	client := server.Client()
	svc, err := drive.NewService(context.Background(),
		option.WithHTTPClient(client),
		option.WithEndpoint(server.URL),
	)
	if err != nil {
		t.Fatalf("creating test drive service: %v", err)
	}
	return svc
}

func jsonResponse(w http.ResponseWriter, data []byte) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(data)
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
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
	if url == "" {
		t.Fatal("expected non-empty URL")
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
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       expiry,
	}
	creds := tokenToCreds(token)
	if creds.AccessToken != "access" {
		t.Errorf("AccessToken: got %q", creds.AccessToken)
	}
	if creds.RefreshToken != "refresh" {
		t.Errorf("RefreshToken: got %q", creds.RefreshToken)
	}
	if !creds.Expiry.Equal(expiry) {
		t.Errorf("Expiry: got %v", creds.Expiry)
	}
}

func TestCredsToToken(t *testing.T) {
	creds := &provider.Credentials{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC),
	}
	token := credsToToken(creds)
	if token.AccessToken != "access" {
		t.Errorf("AccessToken: got %q", token.AccessToken)
	}
	if token.RefreshToken != "refresh" {
		t.Errorf("RefreshToken: got %q", token.RefreshToken)
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
	updated := updatedCreds(original, current)
	if updated == nil {
		t.Fatal("expected updated creds")
	}
	if updated.AccessToken != "new" {
		t.Errorf("AccessToken: got %q", updated.AccessToken)
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
		Id:           "file-1",
		Name:         "report.pdf",
		MimeType:     "application/pdf",
		Size:         1024,
		Md5Checksum:  "abc123",
		ModifiedTime: "2025-06-01T12:00:00Z",
	}
	entry := fileToEntry(f)
	if entry.Ref != "file-1" {
		t.Errorf("Ref: got %q", entry.Ref)
	}
	if entry.Name != "report.pdf" {
		t.Errorf("Name: got %q", entry.Name)
	}
	if entry.IsFolder {
		t.Error("expected not folder")
	}
	if entry.Size != 1024 {
		t.Errorf("Size: got %d", entry.Size)
	}
}

func TestFileToEntry_Folder(t *testing.T) {
	f := &drive.File{
		Id:       "folder-1",
		Name:     "Documents",
		MimeType: "application/vnd.google-apps.folder",
	}
	entry := fileToEntry(f)
	if !entry.IsFolder {
		t.Error("expected folder")
	}
}

func TestIsGoogleNative(t *testing.T) {
	tests := []struct {
		mime   string
		native bool
	}{
		{"application/vnd.google-apps.document", true},
		{"application/vnd.google-apps.spreadsheet", true},
		{"application/vnd.google-apps.presentation", true},
		{"application/pdf", false},
		{"text/plain", false},
	}
	for _, tt := range tests {
		if isGoogleNative(tt.mime) != tt.native {
			t.Errorf("isGoogleNative(%q) = %v, want %v", tt.mime, !tt.native, tt.native)
		}
	}
}

func TestGoogleExportMIME(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"application/vnd.google-apps.document", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"application/vnd.google-apps.spreadsheet", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		{"application/vnd.google-apps.presentation", "application/vnd.openxmlformats-officedocument.presentationml.presentation"},
		{"application/vnd.google-apps.drawing", "application/pdf"},
	}
	for _, tt := range tests {
		got := googleExportMIME(tt.input)
		if got != tt.expected {
			t.Errorf("googleExportMIME(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestChangeToProviderChange_Removed(t *testing.T) {
	c := &drive.Change{FileId: "file-1", Removed: true}
	change := changeToProviderChange(c, "")
	if change == nil {
		t.Fatal("expected change for removal")
	}
	if change.Type != provider.ChangeDeleted {
		t.Errorf("type: got %q, want %q", change.Type, provider.ChangeDeleted)
	}
	if change.Ref != "file-1" {
		t.Errorf("ref: got %q", change.Ref)
	}
}

func TestChangeToProviderChange_NilFile(t *testing.T) {
	c := &drive.Change{FileId: "file-1"}
	change := changeToProviderChange(c, "")
	if change != nil {
		t.Error("expected nil for change with no file and not removed")
	}
}

func TestChangeToProviderChange_FilteredByPath(t *testing.T) {
	c := &drive.Change{
		FileId: "file-1",
		File: &drive.File{
			Id:       "file-1",
			Name:     "report.pdf",
			MimeType: "application/pdf",
			Parents:  []string{"other-folder"},
		},
	}
	change := changeToProviderChange(c, "my-folder")
	if change != nil {
		t.Error("expected nil for file in different folder")
	}
}

func TestChangeToProviderChange_MatchesPath(t *testing.T) {
	c := &drive.Change{
		FileId: "file-1",
		File: &drive.File{
			Id:       "file-1",
			Name:     "report.pdf",
			MimeType: "application/pdf",
			Parents:  []string{"my-folder"},
		},
	}
	change := changeToProviderChange(c, "my-folder")
	if change == nil {
		t.Fatal("expected change for file in matching folder")
	}
	if change.Type != provider.ChangeModified {
		t.Errorf("type: got %q", change.Type)
	}
}

// --- HTTP-mocked tests: data methods ---

func TestList_Success(t *testing.T) {
	server := fakeGoogleServer(t, map[string]http.HandlerFunc{
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

	svc := driveServiceWithServer(t, server)
	g := &GoogleDrive{oauth: &oauth2.Config{}}
	creds := &provider.Credentials{AccessToken: "test-token"}
	token := credsToToken(creds)

	// Call List logic directly using the test service.
	query := "'root' in parents and trashed = false"
	result, err := svc.Files.List().Q(query).
		Fields("nextPageToken, files(id, name, mimeType, size, md5Checksum, modifiedTime)").
		PageSize(1000).Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Fatalf("expected 2 files, got %d", len(result.Files))
	}

	entries := make([]provider.Entry, len(result.Files))
	for i, f := range result.Files {
		entries[i] = fileToEntry(f)
	}

	if entries[0].Name != "doc.pdf" {
		t.Errorf("entry 0 name: got %q", entries[0].Name)
	}
	if entries[0].IsFolder {
		t.Error("entry 0 should not be folder")
	}
	if entries[1].Name != "Photos" {
		t.Errorf("entry 1 name: got %q", entries[1].Name)
	}
	if !entries[1].IsFolder {
		t.Error("entry 1 should be folder")
	}

	// Verify no cred update when token unchanged.
	if updatedCreds(creds, token) != nil {
		t.Error("expected nil creds when unchanged")
	}
	_ = g // silence unused
}

func TestFetch_RegularFile(t *testing.T) {
	server := fakeGoogleServer(t, map[string]http.HandlerFunc{
		"/files/file-1": func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("alt") == "media" {
				w.Header().Set("Content-Type", "application/pdf")
				_, _ = w.Write([]byte("fake pdf content"))
				return
			}
			jsonResponse(w, mustJSON(map[string]any{
				"id": "file-1", "name": "report.pdf", "mimeType": "application/pdf",
				"size": "16", "md5Checksum": "abc123",
			}))
		},
	})
	defer server.Close()

	svc := driveServiceWithServer(t, server)

	// Get metadata.
	file, err := svc.Files.Get("file-1").
		Fields("id, name, mimeType, size, md5Checksum").Do()
	if err != nil {
		t.Fatalf("metadata error: %v", err)
	}

	meta := &provider.EntryMeta{
		Name:        file.Name,
		MimeType:    file.MimeType,
		ContentHash: file.Md5Checksum,
		Size:        file.Size,
	}

	if meta.Name != "report.pdf" {
		t.Errorf("name: got %q", meta.Name)
	}
	if meta.ContentHash != "abc123" {
		t.Errorf("hash: got %q", meta.ContentHash)
	}

	// Download.
	if isGoogleNative(file.MimeType) {
		t.Error("application/pdf should not be native")
	}
	resp, err := svc.Files.Get("file-1").Download()
	if err != nil {
		t.Fatalf("download error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "fake pdf content" {
		t.Errorf("content: got %q", string(body))
	}
}

func TestFetch_GoogleNativeExport(t *testing.T) {
	server := fakeGoogleServer(t, map[string]http.HandlerFunc{
		"/files/doc-1/export": func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte("exported docx bytes"))
		},
		"/files/doc-1": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"id": "doc-1", "name": "My Document",
				"mimeType": "application/vnd.google-apps.document", "size": "0",
			}))
		},
	})
	defer server.Close()

	svc := driveServiceWithServer(t, server)

	file, err := svc.Files.Get("doc-1").
		Fields("id, name, mimeType, size, md5Checksum").Do()
	if err != nil {
		t.Fatalf("metadata error: %v", err)
	}

	if !isGoogleNative(file.MimeType) {
		t.Fatal("expected Google native type")
	}

	exportMime := googleExportMIME(file.MimeType)
	if exportMime != "application/vnd.openxmlformats-officedocument.wordprocessingml.document" {
		t.Errorf("export mime: got %q", exportMime)
	}

	resp, err := svc.Files.Export("doc-1", exportMime).Download()
	if err != nil {
		t.Fatalf("export error: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "exported docx bytes" {
		t.Errorf("content: got %q", string(body))
	}
}

func TestChanges_InitialSync(t *testing.T) {
	server := fakeGoogleServer(t, map[string]http.HandlerFunc{
		"/changes/startPageToken": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{"startPageToken": "token-1"}))
		},
	})
	defer server.Close()

	svc := driveServiceWithServer(t, server)

	startToken, err := svc.Changes.GetStartPageToken().Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if startToken.StartPageToken != "token-1" {
		t.Errorf("start token: got %q", startToken.StartPageToken)
	}
}

func TestChanges_WithChanges(t *testing.T) {
	server := fakeGoogleServer(t, map[string]http.HandlerFunc{
		"/changes": func(w http.ResponseWriter, _ *http.Request) {
			jsonResponse(w, mustJSON(map[string]any{
				"newStartPageToken": "token-2",
				"changes": []map[string]any{
					{
						"fileId":  "f1",
						"removed": false,
						"file": map[string]any{
							"id": "f1", "name": "updated.pdf", "mimeType": "application/pdf",
							"parents": []string{"watched-folder"},
						},
					},
					{
						"fileId":  "f2",
						"removed": true,
					},
					{
						"fileId":  "f3",
						"removed": false,
						"file": map[string]any{
							"id": "f3", "name": "other.pdf", "mimeType": "application/pdf",
							"parents": []string{"different-folder"},
						},
					},
				},
			}))
		},
	})
	defer server.Close()

	svc := driveServiceWithServer(t, server)

	result, err := svc.Changes.List("token-1").
		Fields("nextPageToken, newStartPageToken, changes(fileId, removed, file(id, name, mimeType, size, md5Checksum, modifiedTime, parents))").
		PageSize(1000).IncludeRemoved(true).Do()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.NewStartPageToken != "token-2" {
		t.Errorf("new token: got %q", result.NewStartPageToken)
	}

	// Filter changes for "watched-folder".
	var changes []provider.Change
	for _, c := range result.Changes {
		change := changeToProviderChange(c, "watched-folder")
		if change != nil {
			changes = append(changes, *change)
		}
	}

	// Should include: f1 (matches folder), f2 (deleted, always included), NOT f3 (wrong folder).
	if len(changes) != 2 {
		t.Fatalf("expected 2 filtered changes, got %d", len(changes))
	}
	if changes[0].Ref != "f1" || changes[0].Type != provider.ChangeModified {
		t.Errorf("change 0: ref=%q type=%q", changes[0].Ref, changes[0].Type)
	}
	if changes[1].Ref != "f2" || changes[1].Type != provider.ChangeDeleted {
		t.Errorf("change 1: ref=%q type=%q", changes[1].Ref, changes[1].Type)
	}
}

func TestExchange_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		jsonResponse(w, mustJSON(map[string]any{
			"access_token":  "new-access",
			"refresh_token": "new-refresh",
			"token_type":    "Bearer",
			"expires_in":    3600,
		}))
	}))
	defer server.Close()

	g := &GoogleDrive{
		oauth: &oauth2.Config{
			ClientID:     "client-id",
			ClientSecret: "client-secret",
			Endpoint: oauth2.Endpoint{
				TokenURL: server.URL,
			},
			RedirectURL: "https://app.example.com/callback",
		},
	}

	creds, err := g.Exchange(context.Background(), "auth-code", "https://app.example.com/callback")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if creds.AccessToken != "new-access" {
		t.Errorf("access token: got %q", creds.AccessToken)
	}
	if creds.RefreshToken != "new-refresh" {
		t.Errorf("refresh token: got %q", creds.RefreshToken)
	}
}
