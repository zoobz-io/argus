//go:build testing

package extract

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/proto"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestOCR(t *testing.T) {
	tests := []struct {
		name     string
		mime     string
		data     []byte
		ocrText  string
		wantText string
	}{
		{
			name:     "png image",
			mime:     MimePNG,
			data:     []byte("fake png data"),
			ocrText:  "Extracted from PNG",
			wantText: "Extracted from PNG",
		},
		{
			name:     "jpeg image",
			mime:     MimeJPEG,
			data:     []byte("fake jpeg data"),
			ocrText:  "Extracted from JPEG",
			wantText: "Extracted from JPEG",
		},
		{
			name:     "tiff image",
			mime:     MimeTIFF,
			data:     []byte("fake tiff data"),
			ocrText:  "Extracted from TIFF",
			wantText: "Extracted from TIFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &argustest.MockOCR{
				OnExtractText: func(_ context.Context, in *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
					if in.MimeType != tt.mime {
						t.Errorf("expected mime %q, got %q", tt.mime, in.MimeType)
					}
					return &proto.ExtractTextResponse{Text: tt.ocrText}, nil
				},
			}
			ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

			fn := OCR(tt.mime)
			got, err := fn(ctx, tt.data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantText {
				t.Errorf("got %q, want %q", got, tt.wantText)
			}
		})
	}
}

func TestOCR_Error(t *testing.T) {
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, _ *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			return nil, fmt.Errorf("ocr service unavailable")
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	fn := OCR(MimePNG)
	_, err := fn(ctx, []byte("data"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOCR_PassesDocumentBytes(t *testing.T) {
	input := []byte("specific document bytes")
	mock := &argustest.MockOCR{
		OnExtractText: func(_ context.Context, in *proto.ExtractTextRequest) (*proto.ExtractTextResponse, error) {
			if string(in.Document) != string(input) {
				t.Errorf("expected document bytes %q, got %q", input, in.Document)
			}
			return &proto.ExtractTextResponse{Text: "result"}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithOCR(mock))

	fn := OCR(MimeBMP)
	_, err := fn(ctx, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
