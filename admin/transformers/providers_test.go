//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestProviderToAdminResponse(t *testing.T) {
	p := argustest.NewProvider()
	resp := ProviderToAdminResponse(p)

	if resp.ID != "p1" || resp.TenantID != "t1" || resp.Name != "GDrive" || !resp.Active {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.Type != models.ProviderGoogleDrive {
		t.Errorf("type mismatch: %q", resp.Type)
	}
	if !resp.CreatedAt.Equal(argustest.FixtureTime) {
		t.Error("timestamp not mapped")
	}
}

func TestProvidersToAdminList(t *testing.T) {
	result := &models.OffsetResult[models.Provider]{
		Items:  []*models.Provider{argustest.NewProvider()},
		Offset: 0,
		Total:  1,
	}

	resp := ProvidersToAdminList(result)
	if len(resp.Providers) != 1 || resp.Providers[0].ID != "p1" {
		t.Errorf("providers mismatch: %+v", resp.Providers)
	}
	if resp.Total != 1 {
		t.Errorf("expected Total=1, got %d", resp.Total)
	}
}
