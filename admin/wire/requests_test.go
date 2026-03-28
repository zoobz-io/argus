package wire

import (
	"strings"
	"testing"
)

func TestAdminTenantCreateRequest_Validate_Valid(t *testing.T) {
	req := AdminTenantCreateRequest{Name: "Acme Corp", Slug: "acme-corp"}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestAdminTenantCreateRequest_Validate_MissingName(t *testing.T) {
	req := AdminTenantCreateRequest{Slug: "acme"}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "name") {
		t.Errorf("expected error to mention name, got: %v", err)
	}
}

func TestAdminTenantCreateRequest_Validate_MissingSlug(t *testing.T) {
	req := AdminTenantCreateRequest{Name: "Acme"}
	err := req.Validate()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "slug") {
		t.Errorf("expected error to mention slug, got: %v", err)
	}
}

func TestAdminTenantCreateRequest_Validate_NameTooLong(t *testing.T) {
	req := AdminTenantCreateRequest{Name: strings.Repeat("a", 256), Slug: "acme"}
	if err := req.Validate(); err == nil {
		t.Error("expected error for name exceeding 255 chars")
	}
}

func TestAdminTenantCreateRequest_Validate_SlugTooLong(t *testing.T) {
	req := AdminTenantCreateRequest{Name: "Acme", Slug: strings.Repeat("a", 256)}
	if err := req.Validate(); err == nil {
		t.Error("expected error for slug exceeding 255 chars")
	}
}

func TestAdminTenantCreateRequest_Clone(t *testing.T) {
	orig := AdminTenantCreateRequest{Name: "Acme", Slug: "acme"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}
