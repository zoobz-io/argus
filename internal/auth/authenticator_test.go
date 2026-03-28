package auth

import (
	"net/http"
	"testing"
)

func TestExtractBearerToken_Valid(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
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
	r, _ := http.NewRequest("GET", "/", nil)

	_, err := extractBearerToken(r)
	if err == nil {
		t.Error("expected error for missing header")
	}
}

func TestExtractBearerToken_InvalidFormat(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	r.Header.Set("Authorization", "Basic abc123")

	_, err := extractBearerToken(r)
	if err == nil {
		t.Error("expected error for non-Bearer format")
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
