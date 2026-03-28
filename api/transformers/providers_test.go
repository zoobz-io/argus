//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestProviderToResponse(t *testing.T) {
	p := argustest.NewProvider()
	resp := ProviderToResponse(p)

	if resp.ID != "p1" || resp.Name != "GDrive" || resp.Type != models.ProviderGoogleDrive || !resp.Active {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestProvidersToResponse(t *testing.T) {
	providers := []*models.Provider{argustest.NewProvider(), argustest.NewProvider()}
	providers[1].ID = "p2"

	result := ProvidersToResponse(providers)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[1].ID != "p2" {
		t.Errorf("ID mismatch: %q", result[1].ID)
	}
}

func TestApplyProviderCreate(t *testing.T) {
	req := wire.ProviderCreateRequest{Type: "s3", Name: "My S3", Credentials: "creds"}
	p := &models.Provider{}
	ApplyProviderCreate(req, p)

	if p.Type != "s3" || p.Name != "My S3" || p.Credentials != "creds" {
		t.Errorf("fields not applied: %+v", p)
	}
}
