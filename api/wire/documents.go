package wire

import (
	"context"

	"github.com/zoobz-io/sum"
)

// DocumentResponse is the API wire type for a document.
type DocumentResponse struct {
	CurrentVersionID *string `json:"current_version_id,omitempty" description:"Current version ID"`
	Name             string  `json:"name" description:"Document name" example:"report.pdf"`
	MimeType         string  `json:"mime_type" description:"MIME type" example:"application/pdf"`
	ID               string  `json:"id" description:"Document ID" example:"550e8400-e29b-41d4-a716-446655440000"`
	ProviderID       string  `json:"provider_id" description:"Provider ID" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// OnSend applies boundary rules before sending the response.
func (d *DocumentResponse) OnSend(ctx context.Context) error {
	b := sum.MustUse[sum.Boundary[DocumentResponse]](ctx)
	masked, err := b.Send(ctx, *d)
	if err != nil {
		return err
	}
	*d = masked
	return nil
}

// Clone returns a deep copy of the response.
func (d DocumentResponse) Clone() DocumentResponse {
	c := d
	if d.CurrentVersionID != nil {
		v := *d.CurrentVersionID
		c.CurrentVersionID = &v
	}
	return c
}
