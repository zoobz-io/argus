package extract

import (
	"context"
	"testing"

	argustest "github.com/zoobz-io/argus/testing"
)

func TestODP(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "single page",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:presentation="urn:oasis:names:tc:opendocument:xmlns:presentation:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:presentation>
    <draw:page>
      <draw:frame><draw:text-box>
        <text:p>Slide title</text:p>
        <text:p>Bullet point</text:p>
      </draw:text-box></draw:frame>
    </draw:page>
  </office:presentation></office:body>
</office:document-content>`,
			want: "Slide title\nBullet point",
		},
		{
			name: "multiple pages separated by double newline",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:presentation>
    <draw:page>
      <draw:frame><draw:text-box>
        <text:p>Page 1</text:p>
      </draw:text-box></draw:frame>
    </draw:page>
    <draw:page>
      <draw:frame><draw:text-box>
        <text:p>Page 2</text:p>
      </draw:text-box></draw:frame>
    </draw:page>
  </office:presentation></office:body>
</office:document-content>`,
			want: "Page 1\n\nPage 2",
		},
		{
			name: "empty paragraphs skipped",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:presentation>
    <draw:page>
      <draw:frame><draw:text-box>
        <text:p></text:p>
        <text:p>Content</text:p>
      </draw:text-box></draw:frame>
    </draw:page>
  </office:presentation></office:body>
</office:document-content>`,
			want: "Content",
		},
		{
			name: "unclosed page flushes on EOF",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:presentation>
    <draw:page>
      <draw:frame><draw:text-box>
        <text:p>Page 1</text:p>
      </draw:text-box></draw:frame>
    </draw:page>
    <draw:page>
      <draw:frame><draw:text-box>
        <text:p>Unclosed`,
			want: "Page 1\n\nUnclosed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := argustest.BuildZip(t, argustest.ZipFile{Name: "content.xml", Content: tt.xml})
			got, err := ODP(ctx, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestODP_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	_, err := ODP(ctx, []byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}
