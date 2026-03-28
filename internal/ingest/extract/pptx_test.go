package extract

import (
	"context"
	"fmt"
	"testing"

	argustest "github.com/zoobz-io/argus/testing"
)

func TestPPTX(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		slides []string // XML content for each slide
		want   string
	}{
		{
			name: "single slide",
			slides: []string{
				`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld>
    <p:spTree>
      <p:sp>
        <p:txBody>
          <a:p><a:r><a:t>Slide title</a:t></a:r></a:p>
          <a:p><a:r><a:t>Bullet point</a:t></a:r></a:p>
        </p:txBody>
      </p:sp>
    </p:spTree>
  </p:cSld>
</p:sld>`,
			},
			want: "Slide title\nBullet point",
		},
		{
			name: "multiple slides separated by double newline",
			slides: []string{
				`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld><p:spTree><p:sp><p:txBody>
    <a:p><a:r><a:t>Slide 1</a:t></a:r></a:p>
  </p:txBody></p:sp></p:spTree></p:cSld>
</p:sld>`,
				`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld><p:spTree><p:sp><p:txBody>
    <a:p><a:r><a:t>Slide 2</a:t></a:r></a:p>
  </p:txBody></p:sp></p:spTree></p:cSld>
</p:sld>`,
			},
			want: "Slide 1\n\nSlide 2",
		},
		{
			name: "multiple text boxes on one slide",
			slides: []string{
				`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld><p:spTree>
    <p:sp><p:txBody><a:p><a:r><a:t>Title</a:t></a:r></a:p></p:txBody></p:sp>
    <p:sp><p:txBody><a:p><a:r><a:t>Subtitle</a:t></a:r></a:p></p:txBody></p:sp>
  </p:spTree></p:cSld>
</p:sld>`,
			},
			want: "Title\nSubtitle",
		},
		{
			name: "empty paragraphs skipped",
			slides: []string{
				`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld><p:spTree><p:sp><p:txBody>
    <a:p></a:p>
    <a:p><a:r><a:t>Content</a:t></a:r></a:p>
    <a:p></a:p>
  </p:txBody></p:sp></p:spTree></p:cSld>
</p:sld>`,
			},
			want: "Content",
		},
		{
			name: "unclosed paragraph flushes on EOF",
			slides: []string{
				`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld><p:spTree><p:sp><p:txBody>
    <a:p><a:r><a:t>Closed</a:t></a:r></a:p>
    <a:p><a:r><a:t>Unclosed`,
			},
			want: "Closed\nUnclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var files []argustest.ZipFile
			for i, xml := range tt.slides {
				files = append(files, argustest.ZipFile{
					Name:    fmt.Sprintf("ppt/slides/slide%d.xml", i+1),
					Content: xml,
				})
			}
			data := argustest.BuildZip(t, files...)
			got, err := PPTX(ctx, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPPTX_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	_, err := PPTX(ctx, []byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestPPTX_NoSlides(t *testing.T) {
	ctx := context.Background()
	data := argustest.BuildZip(t, argustest.ZipFile{Name: "ppt/presentation.xml", Content: "<presentation/>"})
	got, err := PPTX(ctx, data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}
