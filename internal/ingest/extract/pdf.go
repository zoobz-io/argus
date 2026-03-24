package extract

import (
	"context"
	"fmt"

	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// PDF extracts text from PDF documents. Currently delegates to OCR which
// handles both text-based and scanned PDFs. A future optimization could
// attempt direct text extraction first and fall back to OCR only for
// image-only PDFs.
func PDF(ctx context.Context, data []byte) (string, error) {
	ocr := sum.MustUse[intcontracts.OCR](ctx)
	resp, err := ocr.ExtractText(ctx, &proto.ExtractTextRequest{
		Document: data,
		MimeType: "application/pdf",
	})
	if err != nil {
		return "", fmt.Errorf("pdf extraction: %w", err)
	}
	return resp.GetText(), nil
}
