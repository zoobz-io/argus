//go:build testing

package wire

import (
	"testing"
)

func TestProviderListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := ProviderListResponse{
		Providers: []ProviderResponse{{ID: "p1", Name: "GDrive"}},
		Limit:     20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProviderListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := ProviderListResponse{}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestProviderListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := ProviderListResponse{
		Providers: []ProviderResponse{{ID: "p1"}},
	}
	cloned := orig.Clone()

	cloned.Providers[0].ID = "mutated"
	if orig.Providers[0].ID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestProviderListResponse_Clone_ValueCopy(t *testing.T) {
	orig := ProviderListResponse{Offset: 10, Limit: 20, Total: 100}
	cloned := orig.Clone()

	cloned.Offset = 999
	if orig.Offset == 999 {
		t.Error("mutating clone affected original")
	}
}
