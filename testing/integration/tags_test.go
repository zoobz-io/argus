//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestTags_CRUD(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "Tag Test", "tag-test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// Create.
	tag, err := s.Tags.CreateTag(ctx, tenant.ID, "compliance", "Compliance content")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}
	if tag.ID == "" || tag.Name != "compliance" {
		t.Errorf("fields mismatch: %+v", tag)
	}

	// Get.
	got, err := s.Tags.GetTag(ctx, tag.ID)
	if err != nil {
		t.Fatalf("get tag: %v", err)
	}
	if got.Name != "compliance" {
		t.Errorf("name mismatch: %q", got.Name)
	}

	// Update.
	updated, err := s.Tags.UpdateTag(ctx, tag.ID, "security", "Security content")
	if err != nil {
		t.Fatalf("update tag: %v", err)
	}
	if updated.Name != "security" {
		t.Errorf("update mismatch: %+v", updated)
	}

	// List by tenant.
	tags, err := s.Tags.ListTagsByTenant(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}

	// Delete.
	if err := s.Tags.DeleteTag(ctx, tag.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
