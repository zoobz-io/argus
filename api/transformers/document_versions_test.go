//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestDocumentVersionToResponse(t *testing.T) {
	v := argustest.NewDocumentVersion()
	resp := DocumentVersionToResponse(v)

	if resp.ID != "v1" || resp.DocumentID != "d1" || resp.VersionNumber != 3 {
		t.Errorf("field mismatch: %+v", resp)
	}
}

func TestDocumentVersionsToResponse(t *testing.T) {
	versions := []*models.DocumentVersion{argustest.NewDocumentVersion(), argustest.NewDocumentVersion()}
	versions[1].ID = "v2"

	result := DocumentVersionsToResponse(versions)
	if len(result) != 2 || result[1].ID != "v2" {
		t.Errorf("unexpected result: %+v", result)
	}
}
