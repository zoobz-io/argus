package auth

import "testing"

func TestZitadelIdentity_ID(t *testing.T) {
	id := &ZitadelIdentity{sub: "user-1"}
	if id.ID() != "user-1" {
		t.Errorf("ID() = %s, want user-1", id.ID())
	}
}

func TestZitadelIdentity_TenantID(t *testing.T) {
	id := &ZitadelIdentity{tenantID: "tenant-1"}
	if id.TenantID() != "tenant-1" {
		t.Errorf("TenantID() = %s, want tenant-1", id.TenantID())
	}
}

func TestZitadelIdentity_Email(t *testing.T) {
	id := &ZitadelIdentity{email: "user@example.com"}
	if id.Email() != "user@example.com" {
		t.Errorf("Email() = %s, want user@example.com", id.Email())
	}
}

func TestZitadelIdentity_Roles(t *testing.T) {
	id := &ZitadelIdentity{roles: []string{"admin", "member"}}
	if len(id.Roles()) != 2 {
		t.Errorf("Roles() length = %d, want 2", len(id.Roles()))
	}
}

func TestZitadelIdentity_Scopes(t *testing.T) {
	id := &ZitadelIdentity{scopes: []string{"openid", "profile"}}
	if len(id.Scopes()) != 2 {
		t.Errorf("Scopes() length = %d, want 2", len(id.Scopes()))
	}
}

func TestZitadelIdentity_HasScope_Found(t *testing.T) {
	id := &ZitadelIdentity{scopes: []string{"openid", "profile"}}
	if !id.HasScope("openid") {
		t.Error("HasScope(openid) = false, want true")
	}
}

func TestZitadelIdentity_HasScope_NotFound(t *testing.T) {
	id := &ZitadelIdentity{scopes: []string{"openid"}}
	if id.HasScope("email") {
		t.Error("HasScope(email) = true, want false")
	}
}

func TestZitadelIdentity_HasScope_Empty(t *testing.T) {
	id := &ZitadelIdentity{}
	if id.HasScope("anything") {
		t.Error("HasScope on empty identity should return false")
	}
}

func TestZitadelIdentity_HasRole_Found(t *testing.T) {
	id := &ZitadelIdentity{roles: []string{"admin", "member"}}
	if !id.HasRole("admin") {
		t.Error("HasRole(admin) = false, want true")
	}
}

func TestZitadelIdentity_HasRole_NotFound(t *testing.T) {
	id := &ZitadelIdentity{roles: []string{"member"}}
	if id.HasRole("admin") {
		t.Error("HasRole(admin) = true, want false")
	}
}

func TestZitadelIdentity_HasRole_Empty(t *testing.T) {
	id := &ZitadelIdentity{}
	if id.HasRole("anything") {
		t.Error("HasRole on empty identity should return false")
	}
}

func TestZitadelIdentity_Stats(t *testing.T) {
	id := &ZitadelIdentity{}
	if id.Stats() != nil {
		t.Error("Stats() should return nil")
	}
}
