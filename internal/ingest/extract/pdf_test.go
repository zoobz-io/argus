//go:build testing

package extract

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/proto"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestPDF_TextLayerExtraction(t *testing.T) {
	// Valid PDF with embedded text — should extract without OCR.
	pdfData := argustest.ReadTestdata(t, "sample.pdf")

	ocrCalled := false
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, _ *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			ocrCalled = true
			return &proto.ExtractTextResponse{Text: "OCR result"}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	got, err := PDF(ctx, pdfData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Hello from PDF" {
		t.Errorf("got %q, want %q", got, "Hello from PDF")
	}
	if ocrCalled {
		t.Error("OCR should not have been called for text-layer PDF")
	}
}

func TestPDF_CorruptStreamFallsBackToOCR(t *testing.T) {
	// Valid PDF structure but corrupt FlateDecode content stream.
	// GetPlainText errors, triggering OCR fallback.
	pdfData := argustest.ReadTestdata(t, "corrupt_stream.pdf")

	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, _ *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			return &proto.ExtractTextResponse{Text: "OCR recovered"}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	got, err := PDF(ctx, pdfData)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "OCR recovered" {
		t.Errorf("got %q, want %q", got, "OCR recovered")
	}
}

func TestPDF_FallsBackToOCR(t *testing.T) {
	// Non-PDF data — text extraction will fail, should fall back to OCR.
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, in *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			if in.MimeType != MimePDF {
				t.Errorf("expected mime %q, got %q", MimePDF, in.MimeType)
			}
			return &proto.ExtractTextResponse{Text: "OCR fallback text"}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	got, err := PDF(ctx, []byte("not a real pdf"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "OCR fallback text" {
		t.Errorf("got %q, want %q", got, "OCR fallback text")
	}
}

func TestPDF_EmptyTextFallsBackToOCR(t *testing.T) {
	// Valid enough for the parser to not error, but no text content.
	// Simulate by providing bytes that parse but yield empty text —
	// the easiest way is a minimal PDF header with no actual content stream.
	minimalPDF := []byte("%PDF-1.0\n%%EOF\n")
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, _ *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			return &proto.ExtractTextResponse{Text: "OCR for empty PDF"}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	got, err := PDF(ctx, minimalPDF)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "OCR for empty PDF" {
		t.Errorf("got %q, want %q", got, "OCR for empty PDF")
	}
}

func TestPDF_OCRFallbackError(t *testing.T) {
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, _ *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			return nil, fmt.Errorf("ocr unavailable")
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	_, err := PDF(ctx, []byte("not a real pdf"))
	if err == nil {
		t.Fatal("expected error when both text extraction and OCR fail")
	}
}

func TestPDF_PassesDataToOCR(t *testing.T) {
	input := []byte("specific pdf bytes")
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, in *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			if string(in.Document) != string(input) {
				t.Errorf("expected document bytes %q, got %q", input, in.Document)
			}
			return &proto.ExtractTextResponse{Text: "result"}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	_, err := PDF(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
