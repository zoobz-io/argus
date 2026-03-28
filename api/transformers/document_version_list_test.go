//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestDocumentVersionsToListResponse(t *testing.T) {
	result := &models.OffsetResult[models.DocumentVersion]{
		Items:  []*models.DocumentVersion{argustest.NewDocumentVersion()},
		Offset: 0,
		Total:  1,
	}

	resp := DocumentVersionsToListResponse(result)
	if len(resp.Versions) != 1 || resp.Versions[0].ID != "v1" {
		t.Errorf("versions mismatch: %+v", resp.Versions)
	}
	if resp.Total != 1 {
		t.Errorf("expected Total=1, got %d", resp.Total)
	}
}
