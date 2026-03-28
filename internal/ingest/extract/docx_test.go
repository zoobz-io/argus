package extract

import (
	"context"
	"testing"

	argustest "github.com/zoobz-io/argus/testing"
)

func TestDOCX(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name  string
		xml   string
		want  string
	}{
		{
			name: "single paragraph",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Hello world</w:t></w:r></w:p>
  </w:body>
</w:document>`,
			want: "Hello world",
		},
		{
			name: "multiple paragraphs",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>First paragraph</w:t></w:r></w:p>
    <w:p><w:r><w:t>Second paragraph</w:t></w:r></w:p>
  </w:body>
</w:document>`,
			want: "First paragraph\nSecond paragraph",
		},
		{
			name: "multiple runs in one paragraph",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p>
      <w:r><w:t>Hello </w:t></w:r>
      <w:r><w:t>world</w:t></w:r>
    </w:p>
  </w:body>
</w:document>`,
			want: "Hello world",
		},
		{
			name: "tab characters",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Column A</w:t><w:tab/><w:t>Column B</w:t></w:r></w:p>
  </w:body>
</w:document>`,
			want: "Column A\tColumn B",
		},
		{
			name: "empty paragraphs are skipped",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p></w:p>
    <w:p><w:r><w:t>Content</w:t></w:r></w:p>
    <w:p></w:p>
  </w:body>
</w:document>`,
			want: "Content",
		},
		{
			name: "line break element",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Before</w:t><w:br/><w:t>After</w:t></w:r></w:p>
  </w:body>
</w:document>`,
			want: "Before\nAfter",
		},
		{
			name: "trailing text without closing paragraph",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Closed</w:t></w:r></w:p>
    <w:r><w:t>Orphan text</w:t></w:r>
  </w:body>
</w:document>`,
			want: "Closed\nOrphan text",
		},
		{
			name: "table cells as paragraphs",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:tbl>
      <w:tr>
        <w:tc><w:p><w:r><w:t>Cell A</w:t></w:r></w:p></w:tc>
        <w:tc><w:p><w:r><w:t>Cell B</w:t></w:r></w:p></w:tc>
      </w:tr>
    </w:tbl>
  </w:body>
</w:document>`,
			want: "Cell A\nCell B",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := argustest.BuildZip(t, argustest.ZipFile{Name: "word/document.xml", Content: tt.xml})
			got, err := DOCX(ctx, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDOCX_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	_, err := DOCX(ctx, []byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestDOCX_MissingDocumentXML(t *testing.T) {
	ctx := context.Background()
	data := argustest.BuildZip(t, argustest.ZipFile{Name: "word/styles.xml", Content: "<styles/>"})
	_, err := DOCX(ctx, data)
	if err == nil {
		t.Fatal("expected error for missing document.xml")
	}
}
