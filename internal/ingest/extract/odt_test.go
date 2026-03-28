package extract

import (
	"context"
	"testing"

	argustest "github.com/zoobz-io/argus/testing"
)

func TestODT(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "single paragraph",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:p>Hello world</text:p>
  </office:text></office:body>
</office:document-content>`,
			want: "Hello world",
		},
		{
			name: "multiple paragraphs",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:p>First</text:p>
    <text:p>Second</text:p>
  </office:text></office:body>
</office:document-content>`,
			want: "First\nSecond",
		},
		{
			name: "headings",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:h>Heading</text:h>
    <text:p>Body text</text:p>
  </office:text></office:body>
</office:document-content>`,
			want: "Heading\nBody text",
		},
		{
			name: "tabs and line breaks",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:p>Col A<text:tab/>Col B</text:p>
    <text:p>Line 1<text:line-break/>Line 2</text:p>
  </office:text></office:body>
</office:document-content>`,
			want: "Col A\tCol B\nLine 1\nLine 2",
		},
		{
			name: "unclosed paragraph flushes on EOF",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:p>Closed</text:p>
    <text:p>Unclosed`,
			want: "Closed\nUnclosed",
		},
		{
			name: "empty paragraphs skipped",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:text>
    <text:p></text:p>
    <text:p>Content</text:p>
    <text:p></text:p>
  </office:text></office:body>
</office:document-content>`,
			want: "Content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := argustest.BuildZip(t, argustest.ZipFile{Name: "content.xml", Content: tt.xml})
			got, err := ODT(ctx, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestODT_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	_, err := ODT(ctx, []byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}

func TestODT_MissingContentXML(t *testing.T) {
	ctx := context.Background()
	data := argustest.BuildZip(t, argustest.ZipFile{Name: "styles.xml", Content: "<styles/>"})
	_, err := ODT(ctx, data)
	if err == nil {
		t.Fatal("expected error for missing content.xml")
	}
}
