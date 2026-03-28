//go:build testing

package wire

import (
	"testing"
)

func TestTenantResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := TenantResponse{ID: "t1", Name: "Acme", Slug: "acme"}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "t1" || resp.Name != "Acme" || resp.Slug != "acme" {
		t.Errorf("fields mutated unexpectedly: %+v", resp)
	}
}

func TestTenantResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := TenantResponse{ID: "t1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestTenantResponse_Clone(t *testing.T) {
	orig := TenantResponse{ID: "t1", Name: "Acme", Slug: "acme"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}
