package auth

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
)

// newTestRequest creates an http.Request with context for linter compliance.
func newTestRequest(t *testing.T, method, url string) *http.Request {
	t.Helper()
	r, err := http.NewRequestWithContext(context.Background(), method, url, nil)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func TestExtractBearerToken_Valid(t *testing.T) {
	r := newTestRequest(t, "GET", "/")
	r.Header.Set("Authorization", "Bearer abc123")

	token, err := extractBearerToken(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "abc123" {
		t.Errorf("expected abc123, got %s", token)
	}
}

func TestExtractBearerToken_Missing(t *testing.T) {
	r := newTestRequest(t, "GET", "/")

	_, err := extractBearerToken(r)
	if err == nil {
		t.Error("expected error for missing header")
	}
}

func TestExtractBearerToken_InvalidFormat(t *testing.T) {
	r := newTestRequest(t, "GET", "/")
	r.Header.Set("Authorization", "Basic abc123")

	_, err := extractBearerToken(r)
	if err == nil {
		t.Error("expected error for non-Bearer format")
	}
}

func TestIdentityFromClaims(t *testing.T) {
	c := &claims{
		Email: "user@example.com",
		OrgID: "org-123",
		Scope: "openid profile email",
		Roles: map[string]map[string]string{
			"member": {"org-123": "org-123"},
			"admin":  {"org-123": "org-123"},
		},
	}

	id := identityFromClaims("user-1", c)

	if id.ID() != "user-1" {
		t.Errorf("ID() = %s, want user-1", id.ID())
	}
	if id.TenantID() != "org-123" {
		t.Errorf("TenantID() = %s, want org-123", id.TenantID())
	}
	if id.Email() != "user@example.com" {
		t.Errorf("Email() = %s, want user@example.com", id.Email())
	}
	// Roles are sorted.
	if len(id.Roles()) != 2 || id.Roles()[0] != "admin" || id.Roles()[1] != "member" {
		t.Errorf("Roles() = %v, want [admin member]", id.Roles())
	}
	if len(id.Scopes()) != 3 {
		t.Errorf("Scopes() length = %d, want 3", len(id.Scopes()))
	}
	if !id.HasScope("openid") {
		t.Error("HasScope(openid) = false, want true")
	}
	if !id.HasRole("admin") {
		t.Error("HasRole(admin) = false, want true")
	}
}

func TestIdentityFromClaims_Empty(t *testing.T) {
	c := &claims{}
	id := identityFromClaims("sub", c)

	if len(id.Roles()) != 0 {
		t.Errorf("expected no roles, got %v", id.Roles())
	}
	if len(id.Scopes()) != 0 {
		t.Errorf("expected no scopes, got %v", id.Scopes())
	}
}

func TestIdentityFromClaims_NoScopes(t *testing.T) {
	c := &claims{
		Scope: "",
		Roles: map[string]map[string]string{"viewer": {}},
	}
	id := identityFromClaims("sub", c)

	if len(id.Scopes()) != 0 {
		t.Errorf("expected no scopes for empty scope string, got %v", id.Scopes())
	}
	if len(id.Roles()) != 1 || id.Roles()[0] != "viewer" {
		t.Errorf("Roles() = %v, want [viewer]", id.Roles())
	}
}

func TestZitadelIdentity_Methods(t *testing.T) {
	id := &ZitadelIdentity{
		sub:      "user-1",
		tenantID: "tenant-1",
		email:    "user@example.com",
		roles:    []string{"admin", "member"},
		scopes:   []string{"openid", "profile"},
	}

	if id.ID() != "user-1" {
		t.Errorf("ID() = %s, want user-1", id.ID())
	}
	if id.TenantID() != "tenant-1" {
		t.Errorf("TenantID() = %s, want tenant-1", id.TenantID())
	}
	if id.Email() != "user@example.com" {
		t.Errorf("Email() = %s, want user@example.com", id.Email())
	}
	if !id.HasRole("admin") {
		t.Error("HasRole(admin) = false, want true")
	}
	if id.HasRole("owner") {
		t.Error("HasRole(owner) = true, want false")
	}
	if !id.HasScope("openid") {
		t.Error("HasScope(openid) = false, want true")
	}
	if id.HasScope("email") {
		t.Error("HasScope(email) = true, want false")
	}
	if id.Stats() != nil {
		t.Error("Stats() should return nil")
	}
	if len(id.Roles()) != 2 {
		t.Errorf("Roles() length = %d, want 2", len(id.Roles()))
	}
	if len(id.Scopes()) != 2 {
		t.Errorf("Scopes() length = %d, want 2", len(id.Scopes()))
	}
}

func TestZitadelIdentity_Empty(t *testing.T) {
	id := &ZitadelIdentity{}

	if id.ID() != "" {
		t.Error("empty ID should return empty string")
	}
	if id.TenantID() != "" {
		t.Error("empty TenantID should return empty string")
	}
	if id.HasRole("anything") {
		t.Error("empty identity should not have any role")
	}
	if id.HasScope("anything") {
		t.Error("empty identity should not have any scope")
	}
}

// fakeOIDCServer creates an httptest server that serves OIDC discovery and JWKS endpoints.
// Returns the server and a function to sign JWTs with the server's key.
func fakeOIDCServer(t *testing.T) (*httptest.Server, func(cl interface{}) string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}

	jwk := jose.JSONWebKey{Key: key.Public(), KeyID: "test-kid", Algorithm: string(jose.RS256), Use: "sig"}
	jwks := jose.JSONWebKeySet{Keys: []jose.JSONWebKey{jwk}}

	mux := http.NewServeMux()
	var issuer string

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"issuer":   issuer,
			"jwks_uri": issuer + "/jwks",
		})
	})
	mux.HandleFunc("/jwks", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(jwks)
	})

	srv := httptest.NewServer(mux)
	issuer = srv.URL

	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, (&jose.SignerOptions{}).WithType("JWT").WithHeader("kid", "test-kid"))
	if err != nil {
		t.Fatal(err)
	}

	sign := func(cl interface{}) string {
		raw, err := jwt.Signed(signer).Claims(cl).Serialize()
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}

	return srv, sign
}

func TestNewAuthenticator_Success(t *testing.T) {
	srv, sign := fakeOIDCServer(t)
	defer srv.Close()

	ctx := context.Background()
	authFn, err := NewAuthenticator(ctx, srv.URL, "")
	if err != nil {
		t.Fatalf("NewAuthenticator failed: %v", err)
	}

	token := sign(map[string]interface{}{
		"iss":   srv.URL,
		"sub":   "user-42",
		"aud":   "test-app",
		"exp":   time.Now().Add(time.Hour).Unix(),
		"iat":   time.Now().Unix(),
		"email": "test@example.com",
		"scope": "openid profile",
		"urn:zitadel:iam:user:resourceowner:id":  "org-99",
		"urn:zitadel:iam:org:project:roles": map[string]interface{}{
			"admin":  map[string]string{"org-99": "org-99"},
			"member": map[string]string{"org-99": "org-99"},
		},
	})

	req := newTestRequest(t, "GET", "/")
	req.Header.Set("Authorization", "Bearer "+token)

	identity, err := authFn(ctx, req)
	if err != nil {
		t.Fatalf("authenticator returned error: %v", err)
	}

	if identity.ID() != "user-42" {
		t.Errorf("ID() = %s, want user-42", identity.ID())
	}
	if identity.TenantID() != "org-99" {
		t.Errorf("TenantID() = %s, want org-99", identity.TenantID())
	}
	if identity.Email() != "test@example.com" {
		t.Errorf("Email() = %s, want test@example.com", identity.Email())
	}
	if !identity.HasRole("admin") {
		t.Error("HasRole(admin) = false, want true")
	}
	if !identity.HasScope("openid") {
		t.Error("HasScope(openid) = false, want true")
	}
}

func TestNewAuthenticator_MissingToken(t *testing.T) {
	srv, _ := fakeOIDCServer(t)
	defer srv.Close()

	ctx := context.Background()
	authFn, err := NewAuthenticator(ctx, srv.URL, "")
	if err != nil {
		t.Fatalf("NewAuthenticator failed: %v", err)
	}

	req := newTestRequest(t, "GET", "/")
	_, err = authFn(ctx, req)
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestNewAuthenticator_InvalidToken(t *testing.T) {
	srv, _ := fakeOIDCServer(t)
	defer srv.Close()

	ctx := context.Background()
	authFn, err := NewAuthenticator(ctx, srv.URL, "")
	if err != nil {
		t.Fatalf("NewAuthenticator failed: %v", err)
	}

	req := newTestRequest(t, "GET", "/")
	req.Header.Set("Authorization", "Bearer not-a-valid-jwt")
	_, err = authFn(ctx, req)
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestNewAuthenticator_ExpiredToken(t *testing.T) {
	srv, sign := fakeOIDCServer(t)
	defer srv.Close()

	ctx := context.Background()
	authFn, err := NewAuthenticator(ctx, srv.URL, "")
	if err != nil {
		t.Fatalf("NewAuthenticator failed: %v", err)
	}

	token := sign(map[string]interface{}{
		"iss": srv.URL,
		"sub": "user-1",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})

	req := newTestRequest(t, "GET", "/")
	req.Header.Set("Authorization", "Bearer "+token)
	_, err = authFn(ctx, req)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestNewAuthenticator_InvalidIssuer(t *testing.T) {
	ctx := context.Background()
	_, err := NewAuthenticator(ctx, "http://localhost:1/nonexistent", "")
	if err == nil {
		t.Error("expected error for invalid issuer")
	}
}

func TestNewAuthenticator_WithAudience(t *testing.T) {
	srv, sign := fakeOIDCServer(t)
	defer srv.Close()

	ctx := context.Background()
	authFn, err := NewAuthenticator(ctx, srv.URL, "my-app")
	if err != nil {
		t.Fatalf("NewAuthenticator failed: %v", err)
	}

	// Token with correct audience.
	token := sign(map[string]interface{}{
		"iss": srv.URL,
		"sub": "user-1",
		"aud": "my-app",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	req := newTestRequest(t, "GET", "/")
	req.Header.Set("Authorization", "Bearer "+token)
	identity, err := authFn(ctx, req)
	if err != nil {
		t.Fatalf("authenticator returned error: %v", err)
	}
	if identity.ID() != "user-1" {
		t.Errorf("ID() = %s, want user-1", identity.ID())
	}

	// Token with wrong audience.
	badToken := sign(map[string]interface{}{
		"iss": srv.URL,
		"sub": "user-1",
		"aud": "wrong-app",
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	req2 := newTestRequest(t, "GET", "/")
	req2.Header.Set("Authorization", "Bearer "+badToken)
	_, err = authFn(ctx, req2)
	if err == nil {
		t.Error("expected error for wrong audience")
	}
}
