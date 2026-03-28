//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestDocumentsToListResponse(t *testing.T) {
	result := &models.OffsetResult[models.Document]{
		Items:  []*models.Document{argustest.NewDocument()},
		Offset: 0,
		Total:  1,
	}

	resp := DocumentsToListResponse(result)
	if len(resp.Documents) != 1 || resp.Documents[0].ID != "d1" {
		t.Errorf("documents mismatch: %+v", resp.Documents)
	}
	if resp.Total != 1 || resp.Offset != 0 {
		t.Errorf("pagination mismatch: %+v", resp)
	}
}
