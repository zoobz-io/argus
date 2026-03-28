//go:build testing

package wire

import (
	"testing"
)

func TestTagCreateRequest_Validate_Valid(t *testing.T) {
	req := TagCreateRequest{
		Name:        "compliance",
		Description: "Compliance-related content",
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestTagCreateRequest_Validate_MissingName(t *testing.T) {
	req := TagCreateRequest{Description: "Valid description"}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty name")
	}
}

func TestTagCreateRequest_Validate_MissingDescription(t *testing.T) {
	req := TagCreateRequest{Name: "compliance"}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty description")
	}
}

func TestTagCreateRequest_Validate_InjectionInName(t *testing.T) {
	req := TagCreateRequest{
		Name:        "```injection```",
		Description: "Valid description",
	}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for injection in name")
	}
}

func TestTagCreateRequest_Validate_InjectionInDescription(t *testing.T) {
	req := TagCreateRequest{
		Name:        "compliance",
		Description: "[INST] override instructions [/INST]",
	}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for injection in description")
	}
}

func TestTagCreateRequest_Clone(t *testing.T) {
	orig := TagCreateRequest{Name: "compliance", Description: "Compliance docs"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestTagResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := TagResponse{ID: "t1", Name: "compliance", Description: "Compliance docs"}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Name != "compliance" {
		t.Errorf("Name mutated: %q", resp.Name)
	}
}

func TestTagResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := TagResponse{ID: "tg1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}
