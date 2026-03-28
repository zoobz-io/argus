package extract

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ledongthuc/pdf"
	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/proto"
)

// PDF extracts text from PDF documents. Attempts direct text extraction
// first using ledongthuc/pdf, falling back to OCR via the Tesseract
// sidecar for scanned or image-only PDFs.
func PDF(ctx context.Context, data []byte) (string, error) {
	text, err := pdfText(data)
	if err == nil && strings.TrimSpace(text) != "" {
		return strings.TrimSpace(text), nil
	}

	// Text layer empty or unreadable — fall back to OCR.
	ocr := sum.MustUse[intcontracts.OCR](ctx)
	resp, err := ocr.ExtractText(ctx, &proto.ExtractTextRequest{
		Document: data,
		MimeType: MimePDF,
	})
	if err != nil {
		return "", fmt.Errorf("pdf ocr fallback: %w", err)
	}
	return resp.GetText(), nil
}

// pdfText extracts embedded text from a PDF using ledongthuc/pdf.
func pdfText(data []byte) (string, error) {
	r, err := pdf.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("opening pdf: %w", err)
	}

	plainText, err := r.GetPlainText()
	if err != nil {
		return "", fmt.Errorf("extracting text: %w", err)
	}

	text, _ := io.ReadAll(plainText)
	return string(text), nil
}
