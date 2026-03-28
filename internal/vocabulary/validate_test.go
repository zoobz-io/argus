//go:build testing

package vocabulary

import (
	"strings"
	"testing"
)

func TestValidateStage_HappyPath(t *testing.T) {
	stage := newValidateStage()

	vc := &Context{
		TenantID:    "tenant-1",
		Name:        "Finance",
		Description: "Financial topics and reporting.",
	}

	result, err := stage.Process(t.Context(), vc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Name != "Finance" {
		t.Errorf("Name: got %q, want %q", result.Name, "Finance")
	}
	if result.Description != "Financial topics and reporting." {
		t.Errorf("Description: got %q, want %q", result.Description, "Financial topics and reporting.")
	}
}

func TestValidateStage_InvalidName(t *testing.T) {
	stage := newValidateStage()

	vc := &Context{
		TenantID:    "tenant-1",
		Name:        "", // empty — required field
		Description: "A valid description.",
	}

	_, err := stage.Process(t.Context(), vc)
	if err == nil {
		t.Fatal("expected error for invalid name")
	}
	if !strings.Contains(err.Error(), "vocabulary validation failed") {
		t.Errorf("error should mention vocabulary validation failed, got %q", err.Error())
	}
}

func TestValidateStage_InvalidDescription(t *testing.T) {
	stage := newValidateStage()

	vc := &Context{
		TenantID:    "tenant-1",
		Name:        "Finance",
		Description: "", // empty — required field
	}

	_, err := stage.Process(t.Context(), vc)
	if err == nil {
		t.Fatal("expected error for invalid description")
	}
	if !strings.Contains(err.Error(), "vocabulary validation failed") {
		t.Errorf("error should mention vocabulary validation failed, got %q", err.Error())
	}
}
