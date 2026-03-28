//go:build integration

package integration

import (
	"context"
	"testing"
)

func TestTopics_CRUD(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "Topic Test", "topic-test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// Create.
	topic, err := s.Topics.CreateTopic(ctx, tenant.ID, "Security", "Security-related documents")
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}
	if topic.ID == "" || topic.Name != "Security" {
		t.Errorf("fields mismatch: %+v", topic)
	}

	// Get.
	got, err := s.Topics.GetTopic(ctx, topic.ID)
	if err != nil {
		t.Fatalf("get topic: %v", err)
	}
	if got.Name != "Security" {
		t.Errorf("name mismatch: %q", got.Name)
	}

	// Update.
	updated, err := s.Topics.UpdateTopic(ctx, topic.ID, "Finance", "Finance documents")
	if err != nil {
		t.Fatalf("update topic: %v", err)
	}
	if updated.Name != "Finance" {
		t.Errorf("update mismatch: %+v", updated)
	}

	// List by tenant.
	topics, err := s.Topics.ListTopicsByTenant(ctx, tenant.ID)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(topics) != 1 {
		t.Errorf("expected 1 topic, got %d", len(topics))
	}

	// Delete.
	if err := s.Topics.DeleteTopic(ctx, topic.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
}
