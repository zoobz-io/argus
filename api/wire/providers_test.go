//go:build testing

package wire

import (
	"strings"
	"testing"

	"github.com/zoobz-io/argus/models"
)

func TestProviderCreateRequest_Validate_Valid(t *testing.T) {
	req := ProviderCreateRequest{
		Type:        models.ProviderType("google_drive"),
		Name:        "My Google Drive",
		Credentials: `{"token":"abc"}`,
	}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestProviderCreateRequest_Validate_MissingFields(t *testing.T) {
	tests := []struct {
		name  string
		req   ProviderCreateRequest
		field string
	}{
		{
			name:  "missing type",
			req:   ProviderCreateRequest{Name: "test", Credentials: "creds"},
			field: "type",
		},
		{
			name:  "missing name",
			req:   ProviderCreateRequest{Type: "google_drive", Credentials: "creds"},
			field: "name",
		},
		{
			name:  "missing credentials",
			req:   ProviderCreateRequest{Type: "google_drive", Name: "test"},
			field: "credentials",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if err == nil {
				t.Fatal("expected validation error")
			}
			if !strings.Contains(err.Error(), tt.field) {
				t.Errorf("expected error to mention %q, got: %v", tt.field, err)
			}
		})
	}
}

func TestProviderCreateRequest_Validate_NameTooLong(t *testing.T) {
	req := ProviderCreateRequest{
		Type:        "google_drive",
		Name:        strings.Repeat("a", 256),
		Credentials: "creds",
	}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for name exceeding 255 chars")
	}
}

func TestProviderCreateRequest_Clone(t *testing.T) {
	orig := ProviderCreateRequest{Type: "google_drive", Name: "GDrive", Credentials: "creds"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestProviderResponse_Clone(t *testing.T) {
	orig := ProviderResponse{ID: "p1", Name: "GDrive", Type: "google_drive", Active: true}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestProviderResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := ProviderResponse{ID: "p1", Name: "GDrive", Type: "google_drive", Active: true}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "p1" {
		t.Errorf("ID mutated: %q", resp.ID)
	}
}

func TestProviderResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := ProviderResponse{ID: "p1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}
