package extract

import (
	"context"
	"fmt"

	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// OCR extracts text from image and scanned document formats via the
// Tesseract gRPC sidecar. The mimeType parameter is curried at registration
// time so the extractor signature stays consistent.
func OCR(mimeType string) Func {
	return func(ctx context.Context, data []byte) (string, error) {
		ocr := sum.MustUse[intcontracts.OCR](ctx)
		resp, err := ocr.ExtractText(ctx, &proto.ExtractTextRequest{
			Document: data,
			MimeType: mimeType,
		})
		if err != nil {
			return "", fmt.Errorf("ocr extraction: %w", err)
		}
		return resp.GetText(), nil
	}
}
