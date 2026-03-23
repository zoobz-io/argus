package wire

import (
	"context"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/sum"
)

// DocumentVersionResponse is the API wire type for a document version.
type DocumentVersionResponse struct {
	ExtractionStatus models.ExtractionStatus `json:"extraction_status" description:"Content extraction status" example:"completed"`
	ID               int64                   `json:"id" description:"Version ID" example:"1"`
	DocumentID       int64                   `json:"document_id" description:"Document ID" example:"1"`
	VersionNumber    int                     `json:"version_number" description:"Version number" example:"1"`
}

// OnSend applies boundary rules before sending the response.
func (v *DocumentVersionResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[*sum.Boundary[DocumentVersionResponse]](ctx)
	masked, err := b.Send(ctx, *v)
	if err != nil {
		return err
	}
	*v = masked
	return nil
}

// Clone returns a deep copy of the response.
func (v DocumentVersionResponse) Clone() DocumentVersionResponse {
	return v
}
