//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestDocumentToResponse(t *testing.T) {
	d := argustest.NewDocument()
	resp := DocumentToResponse(d)

	if resp.ID != "d1" || resp.Name != "report.pdf" || resp.MimeType != "application/pdf" || resp.ProviderID != "p1" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.CurrentVersionID == nil || *resp.CurrentVersionID != "v1" {
		t.Error("CurrentVersionID not mapped")
	}
}

func TestDocumentsToResponse(t *testing.T) {
	docs := []*models.Document{argustest.NewDocument(), argustest.NewDocument()}
	docs[1].ID = "d2"

	result := DocumentsToResponse(docs)
	if len(result) != 2 || result[1].ID != "d2" {
		t.Errorf("unexpected result: %+v", result)
	}
}
