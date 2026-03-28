package extract

import (
	"context"
	"fmt"

	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// OCR extracts text from image and scanned document formats via the
// Tesseract gRPC sidecar. The mimeType parameter is curried at registration
// time so the extractor signature stays consistent.
func OCR(mimeType string) Func {
	return func(ctx context.Context, data []byte) (string, error) {
		ocr := sum.MustUse[intcontracts.OCR](ctx)
		cfg := sum.MustUse[config.OCR](ctx)
		callCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		resp, err := ocr.ExtractText(callCtx, &proto.ExtractTextRequest{
			Document: data,
			MimeType: mimeType,
		})
		if err != nil {
			return "", fmt.Errorf("ocr extraction: %w", err)
		}
		return resp.GetText(), nil
	}
}
