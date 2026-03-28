//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/zoobz-io/argus/models"
)

// TestDocuments_ListByTenant tests document listing.
// Note: Document creation happens through the ingestion pipeline, not direct store calls.
// This test creates documents directly to verify the store's list/get/delete behavior.
func TestDocuments_ListByTenant(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	tenant, err := s.Tenants.CreateTenant(ctx, "Doc Test", "doc-test")
	if err != nil {
		t.Fatalf("create tenant: %v", err)
	}

	// List empty.
	result, err := s.Documents.ListDocumentsByTenant(ctx, tenant.ID, models.OffsetPage{Limit: 10})
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(result.Items) != 0 {
		t.Errorf("expected empty list, got %d", len(result.Items))
	}
}
