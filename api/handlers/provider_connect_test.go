//go:build testing

package handlers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestIsOAuthProvider(t *testing.T) {
	tests := []struct {
		pt   models.ProviderType
		want bool
	}{
		{models.ProviderGoogleDrive, true},
		{models.ProviderOneDrive, true},
		{models.ProviderDropbox, true},
		{models.ProviderS3, false},
		{models.ProviderGCS, false},
		{models.ProviderAzureBlob, false},
	}
	for _, tt := range tests {
		t.Run(string(tt.pt), func(t *testing.T) {
			if got := isOAuthProvider(tt.pt); got != tt.want {
				t.Errorf("isOAuthProvider(%q) = %v, want %v", tt.pt, got, tt.want)
			}
		})
	}
}

func TestAllowedRedirectURIs_IsAllowed(t *testing.T) {
	a := NewAllowedRedirectURIs("https://app.example.com/callback,https://localhost:3000/cb")

	if !a.IsAllowed("https://app.example.com/callback") {
		t.Error("expected match for allowed URI")
	}
	if a.IsAllowed("https://evil.com/callback") {
		t.Error("expected no match for disallowed URI")
	}

	empty := NewAllowedRedirectURIs("")
	if empty.IsAllowed("https://anything.com") {
		t.Error("expected no match for empty allow list")
	}
}

func TestNewAllowedRedirectURIs(t *testing.T) {
	a := NewAllowedRedirectURIs("  https://a.com , https://b.com , https://c.com  ")
	if !a.IsAllowed("https://a.com") {
		t.Error("expected https://a.com to be allowed")
	}
	if !a.IsAllowed("https://b.com") {
		t.Error("expected https://b.com to be allowed")
	}
	if !a.IsAllowed("https://c.com") {
		t.Error("expected https://c.com to be allowed")
	}
	if len(a.uris) != 3 {
		t.Errorf("expected 3 URIs, got %d", len(a.uris))
	}
}
