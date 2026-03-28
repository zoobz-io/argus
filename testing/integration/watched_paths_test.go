//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestWatchedPaths_CRUD(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "WP Test", "wp-test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	provider, err := s.Providers.CreateProvider(ctx, tenant.ID, models.ProviderS3, "S3", `{}`)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	// Create.
	wp, err := s.WatchedPaths.CreateWatchedPath(ctx, tenant.ID, provider.ID, "/documents/reports")
	if err != nil {
		t.Fatalf("create watched path: %v", err)
	}
	if wp.ID == "" || wp.Path != "/documents/reports" {
		t.Errorf("fields mismatch: %+v", wp)
	}

	// Get.
	got, err := s.WatchedPaths.GetWatchedPath(ctx, wp.ID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Path != "/documents/reports" {
		t.Errorf("path mismatch: %q", got.Path)
	}

	// Update.
	updated, err := s.WatchedPaths.UpdateWatchedPath(ctx, wp.ID, "/documents/archive")
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Path != "/documents/archive" {
		t.Errorf("update mismatch: %+v", updated)
	}

	// List by tenant.
	result, err := s.WatchedPaths.ListWatchedPathsByTenant(ctx, tenant.ID, models.OffsetPage{Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) != 1 {
		t.Errorf("expected 1 watched path, got %d", len(result.Items))
	}

	// Delete.
	if err := s.WatchedPaths.DeleteWatchedPath(ctx, wp.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
