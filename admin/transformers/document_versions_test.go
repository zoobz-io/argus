//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestDocumentVersionToAdminResponse(t *testing.T) {
	v := argustest.NewDocumentVersion()
	resp := DocumentVersionToAdminResponse(v)

	if resp.ID != "v1" || resp.DocumentID != "d1" || resp.TenantID != "t1" || resp.VersionNumber != 3 || resp.ContentHash != "abc123" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if !resp.CreatedAt.Equal(argustest.FixtureTime) {
		t.Error("timestamp not mapped")
	}
}

func TestDocumentVersionsToAdminList(t *testing.T) {
	result := &models.OffsetResult[models.DocumentVersion]{
		Items:  []*models.DocumentVersion{argustest.NewDocumentVersion()},
		Offset: 0,
		Total:  1,
	}

	resp := DocumentVersionsToAdminList(result)
	if len(resp.Versions) != 1 || resp.Versions[0].ID != "v1" {
		t.Errorf("versions mismatch: %+v", resp.Versions)
	}
	if resp.Total != 1 {
		t.Errorf("expected Total=1, got %d", resp.Total)
	}
}
