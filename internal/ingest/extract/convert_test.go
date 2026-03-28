//go:build testing

package extract

import (
	"context"
	"fmt"
	"testing"

	"github.com/zoobz-io/argus/proto"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestConvert_Error(t *testing.T) {
	mock := &argustest.MockConverter{
		OnConvertDocument: func(_ context.Context, _ *proto.ConvertRequest) (*proto.ConvertResponse, error) {
			return nil, fmt.Errorf("conversion failed")
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithConverter(mock))

	fn := Convert(MimeDOC)
	_, err := fn(ctx, []byte("fake doc data"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConvert_UnexpectedOutputMime(t *testing.T) {
	mock := &argustest.MockConverter{
		OnConvertDocument: func(_ context.Context, in *proto.ConvertRequest) (*proto.ConvertResponse, error) {
			return &proto.ConvertResponse{
				Document: []byte("converted data"),
				MimeType: "application/unknown",
			}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithConverter(mock))

	fn := Convert(MimeDOC)
	_, err := fn(ctx, []byte("fake doc data"))
	if err == nil {
		t.Fatal("expected error for unexpected MIME type")
	}
}

func TestConvert_PassesMimeType(t *testing.T) {
	var receivedMime string
	mock := &argustest.MockConverter{
		OnConvertDocument: func(_ context.Context, in *proto.ConvertRequest) (*proto.ConvertResponse, error) {
			receivedMime = in.MimeType
			return nil, fmt.Errorf("stop here")
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithConverter(mock))

	fn := Convert(MimeXLS)
	_, _ = fn(ctx, []byte("fake xls data"))

	if receivedMime != MimeXLS {
		t.Errorf("expected mime %q, got %q", MimeXLS, receivedMime)
	}
}

func TestConvert_RoutesToDOCX(t *testing.T) {
	docxXML := `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body><w:p><w:r><w:t>Converted doc</w:t></w:r></w:p></w:body>
</w:document>`
	docxBytes := argustest.BuildZip(t, argustest.ZipFile{Name: "word/document.xml", Content: docxXML})

	mock := &argustest.MockConverter{
		OnConvertDocument: func(_ context.Context, _ *proto.ConvertRequest) (*proto.ConvertResponse, error) {
			return &proto.ConvertResponse{Document: docxBytes, MimeType: MimeDOCX}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithConverter(mock))

	fn := Convert(MimeDOC)
	got, err := fn(ctx, []byte("fake doc data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Converted doc" {
		t.Errorf("got %q, want %q", got, "Converted doc")
	}
}

func TestConvert_RoutesToXLSX(t *testing.T) {
	xlsxBytes := buildXLSX(t, map[string][][]string{
		"Data": {{"Cell A1"}},
	})

	mock := &argustest.MockConverter{
		OnConvertDocument: func(_ context.Context, _ *proto.ConvertRequest) (*proto.ConvertResponse, error) {
			return &proto.ConvertResponse{Document: xlsxBytes, MimeType: MimeXLSX}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithConverter(mock))

	fn := Convert(MimeXLS)
	got, err := fn(ctx, []byte("fake xls data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Cell A1" {
		t.Errorf("got %q, want %q", got, "Cell A1")
	}
}

func TestConvert_RoutesToPPTX(t *testing.T) {
	slideXML := `<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main"
       xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main">
  <p:cSld><p:spTree>
    <p:sp><p:txBody><a:p><a:r><a:t>Slide text</a:t></a:r></a:p></p:txBody></p:sp>
  </p:spTree></p:cSld>
</p:sld>`
	pptxBytes := argustest.BuildZip(t,
		argustest.ZipFile{Name: "ppt/slides/slide1.xml", Content: slideXML},
	)

	mock := &argustest.MockConverter{
		OnConvertDocument: func(_ context.Context, _ *proto.ConvertRequest) (*proto.ConvertResponse, error) {
			return &proto.ConvertResponse{Document: pptxBytes, MimeType: MimePPTX}, nil
		},
	}
	ctx := argustest.SetupRegistry(t, argustest.WithConverter(mock))

	fn := Convert(MimePPT)
	got, err := fn(ctx, []byte("fake ppt data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "Slide text" {
		t.Errorf("got %q, want %q", got, "Slide text")
	}
}
