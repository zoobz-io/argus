//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestProviders_CRUD(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Create tenant first.
	tenant, err := s.Tenants.CreateTenant(ctx, "Provider Test", "provider-test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// Create provider.
	provider, err := s.Providers.CreateProvider(ctx, tenant.ID, models.ProviderS3, "My S3", `{"key":"value"}`)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	if provider.ID == "" || provider.Name != "My S3" || provider.Type != models.ProviderS3 {
		t.Errorf("fields mismatch: %+v", provider)
	}

	// Get.
	got, err := s.Providers.GetProvider(ctx, provider.ID)
	if err != nil {
		t.Fatalf("get provider: %v", err)
	}
	if got.ID != provider.ID {
		t.Errorf("ID mismatch: %q vs %q", got.ID, provider.ID)
	}

	// Update.
	updated, err := s.Providers.UpdateProvider(ctx, provider.ID, models.ProviderGCS, "My GCS", `{"key":"new"}`)
	if err != nil {
		t.Fatalf("update provider: %v", err)
	}
	if updated.Name != "My GCS" || updated.Type != models.ProviderGCS {
		t.Errorf("update mismatch: %+v", updated)
	}

	// List by tenant.
	result, err := s.Providers.ListProvidersByTenant(ctx, tenant.ID, models.OffsetPage{Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 provider, got %d", len(result.Items))
	}

	// Delete.
	if err := s.Providers.DeleteProvider(ctx, provider.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
