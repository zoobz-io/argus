package googledrive

import (
	"context"
	"testing"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"

	"github.com/zoobz-io/argus/provider"
)

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
	// Should contain our client ID and state.
	if !contains(url, "client_id=client-id") {
		t.Errorf("URL should contain client_id, got %s", url)
	}
	if !contains(url, "state=csrf-token") {
		t.Errorf("URL should contain state, got %s", url)
	}
	if !contains(url, "access_type=offline") {
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

func contains(s, substr string) bool {
	return len(s) >= len(substr) && containsSubstr(s, substr)
}

func containsSubstr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
