//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestProvidersToListResponse(t *testing.T) {
	result := &models.OffsetResult[models.Provider]{
		Items:  []*models.Provider{argustest.NewProvider()},
		Offset: 0,
		Total:  1,
	}

	resp := ProvidersToListResponse(result)
	if len(resp.Providers) != 1 || resp.Providers[0].ID != "p1" {
		t.Errorf("providers mismatch: %+v", resp.Providers)
	}
	if resp.Total != 1 || resp.Offset != 0 {
		t.Errorf("pagination mismatch: offset=%d total=%d", resp.Offset, resp.Total)
	}
}

func TestProvidersToListResponse_Empty(t *testing.T) {
	result := &models.OffsetResult[models.Provider]{Items: []*models.Provider{}}
	resp := ProvidersToListResponse(result)
	if len(resp.Providers) != 0 {
		t.Errorf("expected empty, got %d", len(resp.Providers))
	}
}
