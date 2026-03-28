package extract

import (
	"context"
	"fmt"

	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/config"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// Convert returns an extractor that converts legacy binary Office formats
// to their modern OOXML equivalents via the LibreOffice sidecar, then
// delegates to the appropriate OOXML extractor.
func Convert(inputMime string) Func {
	return func(ctx context.Context, data []byte) (string, error) {
		converter := sum.MustUse[intcontracts.Converter](ctx)
		cfg := sum.MustUse[config.Convert](ctx)
		callCtx, cancel := context.WithTimeout(ctx, cfg.Timeout)
		defer cancel()
		resp, err := converter.ConvertDocument(callCtx, &proto.ConvertRequest{
			Document: data,
			MimeType: inputMime,
		})
		if err != nil {
			return "", fmt.Errorf("converting %s: %w", inputMime, err)
		}

		// Route the converted bytes to the appropriate OOXML extractor.
		switch resp.MimeType {
		case MimeDOCX:
			return DOCX(ctx, resp.Document)
		case MimeXLSX:
			return XLSX(ctx, resp.Document)
		case MimePPTX:
			return PPTX(ctx, resp.Document)
		default:
			return "", fmt.Errorf("unexpected conversion output MIME: %s", resp.MimeType)
		}
	}
}
