//go:build testing

package handlers

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/internal/oauth"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
	argustest "github.com/zoobz-io/argus/testing"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// --- helpers ---

type stubProvider struct {
	authURL string
}

func (s *stubProvider) Type() string { return "google_drive" }
func (s *stubProvider) AuthURL(_ context.Context, _, _ string) (string, error) {
	return s.authURL, nil
}
func (s *stubProvider) Exchange(_ context.Context, _, _ string) (*provider.Credentials, error) {
	return &provider.Credentials{AccessToken: "test-token"}, nil
}
func (s *stubProvider) List(_ context.Context, _ *provider.Credentials, _ string) ([]provider.Entry, *provider.Credentials, error) {
	return nil, nil, nil
}
func (s *stubProvider) Changes(_ context.Context, _ *provider.Credentials, _, _ string) ([]provider.Change, string, *provider.Credentials, error) {
	return nil, "", nil, nil
}
func (s *stubProvider) Fetch(_ context.Context, _ *provider.Credentials, _ string) (io.ReadCloser, *provider.EntryMeta, *provider.Credentials, error) {
	return nil, nil, nil, nil
}

func testRegistry() *provider.Registry {
	reg := provider.NewRegistry()
	reg.Register(&stubProvider{authURL: "https://accounts.google.com/o/oauth2/auth?test=1"})
	return reg
}

func testSigner() *oauth.StateSigner {
	return oauth.NewStateSigner([]byte("test-key-32-bytes-long-for-hmac!"))
}

func testAllowedURIs() *AllowedRedirectURIs {
	return NewAllowedRedirectURIs("http://localhost:3000/callback,https://app.example.com/callback")
}

// --- unit tests ---

func TestIsOAuthProvider(t *testing.T) {
	tests := []struct {
		pt     models.ProviderType
		expect bool
	}{
		{models.ProviderGoogleDrive, true},
		{models.ProviderOneDrive, true},
		{models.ProviderDropbox, true},
		{models.ProviderS3, false},
		{models.ProviderGCS, false},
		{models.ProviderAzureBlob, false},
	}
	for _, tt := range tests {
		if got := isOAuthProvider(tt.pt); got != tt.expect {
			t.Errorf("isOAuthProvider(%q) = %v, want %v", tt.pt, got, tt.expect)
		}
	}
}

func TestAllowedRedirectURIs_IsAllowed(t *testing.T) {
	a := NewAllowedRedirectURIs("http://localhost:3000/callback,https://app.example.com/callback")
	if !a.IsAllowed("http://localhost:3000/callback") {
		t.Error("expected allowed")
	}
	if a.IsAllowed("http://evil.com/callback") {
		t.Error("expected not allowed")
	}
}

func TestNewAllowedRedirectURIs(t *testing.T) {
	a := NewAllowedRedirectURIs("  http://a.com , http://b.com  ,  ")
	if len(a.uris) != 2 {
		t.Errorf("expected 2 URIs, got %d", len(a.uris))
	}
}

func TestAllowedRedirectURIs_Empty(t *testing.T) {
	a := NewAllowedRedirectURIs("")
	if a.IsAllowed("http://anything.com") {
		t.Error("empty list should not allow anything")
	}
}

// --- handler tests ---

func TestProviderAuthURL_Success(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProviderByTenant: func(_ context.Context, id, _ string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: models.ProviderGoogleDrive, TenantID: "tenant-1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(mock),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/providers/p-1/auth-url?redirect_uri=http://localhost:3000/callback", nil)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.AuthURLResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.URL == "" {
		t.Error("expected non-empty URL")
	}
}

func TestProviderAuthURL_MissingRedirectURI(t *testing.T) {
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(&argustest.MockProviders{}),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/providers/p-1/auth-url", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error for missing redirect_uri")
	}
}

func TestProviderAuthURL_InvalidRedirectURI(t *testing.T) {
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(&argustest.MockProviders{}),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/providers/p-1/auth-url?redirect_uri=http://evil.com", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error for disallowed redirect_uri")
	}
}

func TestProviderAuthURL_ProviderNotFound(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProviderByTenant: func(_ context.Context, _, _ string) (*models.Provider, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(mock),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	capture := rtesting.ServeRequest(engine, "GET", "/providers/bad-id/auth-url?redirect_uri=http://localhost:3000/callback", nil)
	if capture.StatusCode() == 200 {
		t.Error("expected error for unknown provider")
	}
}

func TestProviderConnect_OAuth_Success(t *testing.T) {
	signer := testSigner()
	state, _ := signer.Sign("p-1", "tenant-1")

	mock := &argustest.MockProviders{
		OnGetProviderByTenant: func(_ context.Context, id, _ string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: models.ProviderGoogleDrive, TenantID: "tenant-1"}, nil
		},
		OnUpdateProviderCredentials: func(_ context.Context, _, _, _ string) error {
			return nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(mock),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(signer),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	body := wire.ConnectRequest{Code: "auth-code", RedirectURI: "http://localhost:3000/callback", State: state}
	capture := rtesting.ServeRequest(engine, "POST", "/providers/p-1/connect", body)
	rtesting.AssertStatus(t, capture, 200)

	var resp wire.ConnectResponse
	if err := capture.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Status != "connected" {
		t.Errorf("status: got %q", resp.Status)
	}
}

func TestProviderConnect_InvalidState(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProviderByTenant: func(_ context.Context, id, _ string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: models.ProviderGoogleDrive, TenantID: "tenant-1"}, nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(mock),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	body := wire.ConnectRequest{Code: "auth-code", RedirectURI: "http://localhost:3000/callback", State: "bad-state"}
	capture := rtesting.ServeRequest(engine, "POST", "/providers/p-1/connect", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error for invalid state")
	}
}

func TestProviderConnect_StaticCredentials(t *testing.T) {
	creds := `{"AccessToken":"AKID","RefreshToken":"secret"}`
	mock := &argustest.MockProviders{
		OnGetProviderByTenant: func(_ context.Context, id, _ string) (*models.Provider, error) {
			return &models.Provider{ID: id, Type: models.ProviderS3, TenantID: "tenant-1"}, nil
		},
		OnUpdateProviderCredentials: func(_ context.Context, _, _, gotCreds string) error {
			if gotCreds != creds {
				t.Errorf("creds: got %q, want %q", gotCreds, creds)
			}
			return nil
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(mock),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	body := wire.ConnectRequest{Code: creds, RedirectURI: "unused", State: "unused"}
	capture := rtesting.ServeRequest(engine, "POST", "/providers/p-1/connect", body)
	rtesting.AssertStatus(t, capture, 200)
}

func TestProviderConnect_ProviderNotFound(t *testing.T) {
	mock := &argustest.MockProviders{
		OnGetProviderByTenant: func(_ context.Context, _, _ string) (*models.Provider, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	engine := argustest.SetupAPIEngine(t, All(),
		argustest.WithAPIProviders(mock),
		argustest.WithRegistration(func(k sum.Key) { sum.Register[*AllowedRedirectURIs](k, testAllowedURIs()) }),
		argustest.WithProviderRegistry(testRegistry()),
		argustest.WithStateSigner(testSigner()),
		argustest.WithBoundaries(wire.RegisterBoundaries),
	)

	body := wire.ConnectRequest{Code: "code", RedirectURI: "uri", State: "state"}
	capture := rtesting.ServeRequest(engine, "POST", "/providers/p-1/connect", body)
	if capture.StatusCode() == 200 {
		t.Error("expected error for unknown provider")
	}
}
