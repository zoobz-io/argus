package extract

import (
	"context"
	"fmt"

	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// Legacy extracts text from legacy binary Office formats (DOC, XLS, PPT)
// by delegating to OCR. These formats use complex binary structures that
// are impractical to parse directly; OCR provides reliable text extraction.
func Legacy(mimeType string) Func {
	return func(ctx context.Context, data []byte) (string, error) {
		ocr := sum.MustUse[intcontracts.OCR](ctx)
		resp, err := ocr.ExtractText(ctx, &proto.ExtractTextRequest{
			Document: data,
			MimeType: mimeType,
		})
		if err != nil {
			return "", fmt.Errorf("legacy format extraction: %w", err)
		}
		return resp.GetText(), nil
	}
}
