//go:build testing

package vocabulary

import "testing"

func TestContext_Clone_CopyIndependence(t *testing.T) {
	vc := &Context{
		TenantID:    "tenant-1",
		Name:        "Finance",
		Description: "Financial topics",
		ID:          "vocab-1",
	}

	clone := vc.Clone()

	// Verify all fields copied.
	if clone.TenantID != vc.TenantID {
		t.Errorf("TenantID: got %q, want %q", clone.TenantID, vc.TenantID)
	}
	if clone.Name != vc.Name {
		t.Errorf("Name: got %q, want %q", clone.Name, vc.Name)
	}
	if clone.Description != vc.Description {
		t.Errorf("Description: got %q, want %q", clone.Description, vc.Description)
	}
	if clone.ID != vc.ID {
		t.Errorf("ID: got %q, want %q", clone.ID, vc.ID)
	}

	// Verify independence — mutate clone, original unchanged.
	clone.TenantID = "tenant-mutated"
	clone.Name = "Mutated"
	clone.Description = "Mutated desc"
	clone.ID = "mutated-id"

	if vc.TenantID != "tenant-1" {
		t.Error("TenantID mutation leaked to original")
	}
	if vc.Name != "Finance" {
		t.Error("Name mutation leaked to original")
	}
	if vc.Description != "Financial topics" {
		t.Error("Description mutation leaked to original")
	}
	if vc.ID != "vocab-1" {
		t.Error("ID mutation leaked to original")
	}
}

func TestContext_Clone_EmptyFields(t *testing.T) {
	vc := &Context{}
	clone := vc.Clone()

	if clone.TenantID != "" {
		t.Errorf("TenantID: expected empty, got %q", clone.TenantID)
	}
	if clone.Name != "" {
		t.Errorf("Name: expected empty, got %q", clone.Name)
	}
	if clone.Description != "" {
		t.Errorf("Description: expected empty, got %q", clone.Description)
	}
	if clone.ID != "" {
		t.Errorf("ID: expected empty, got %q", clone.ID)
	}
}
