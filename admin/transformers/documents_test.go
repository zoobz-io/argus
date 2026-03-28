//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestDocumentToAdminResponse(t *testing.T) {
	d := argustest.NewDocument()
	resp := DocumentToAdminResponse(d)

	if resp.ID != "d1" || resp.TenantID != "t1" || resp.Name != "report.pdf" || resp.ExternalID != "ext1" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.CurrentVersionID == nil || *resp.CurrentVersionID != "v1" {
		t.Error("CurrentVersionID not mapped")
	}
}

func TestDocumentsToAdminList(t *testing.T) {
	result := &models.OffsetResult[models.Document]{
		Items:  []*models.Document{argustest.NewDocument()},
		Offset: 0,
		Total:  1,
	}

	resp := DocumentsToAdminList(result)
	if len(resp.Documents) != 1 || resp.Documents[0].ID != "d1" {
		t.Errorf("documents mismatch: %+v", resp.Documents)
	}
	if resp.Total != 1 || resp.Offset != 0 {
		t.Errorf("pagination mismatch: %+v", resp)
	}
}
