package extract

import (
	"context"
	"testing"

	argustest "github.com/zoobz-io/argus/testing"
)

func TestODS(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		xml  string
		want string
	}{
		{
			name: "simple table",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:spreadsheet>
    <table:table>
      <table:table-row>
        <table:table-cell><text:p>A1</text:p></table:table-cell>
        <table:table-cell><text:p>B1</text:p></table:table-cell>
      </table:table-row>
      <table:table-row>
        <table:table-cell><text:p>A2</text:p></table:table-cell>
        <table:table-cell><text:p>B2</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet></office:body>
</office:document-content>`,
			want: "A1\tB1\nA2\tB2",
		},
		{
			name: "empty rows skipped",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:spreadsheet>
    <table:table>
      <table:table-row>
        <table:table-cell><text:p>Data</text:p></table:table-cell>
      </table:table-row>
      <table:table-row>
        <table:table-cell></table:table-cell>
      </table:table-row>
      <table:table-row>
        <table:table-cell><text:p>More data</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet></office:body>
</office:document-content>`,
			want: "Data\nMore data",
		},
		{
			name: "single cell",
			xml: `<?xml version="1.0" encoding="UTF-8"?>
<office:document-content xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0">
  <office:body><office:spreadsheet>
    <table:table>
      <table:table-row>
        <table:table-cell><text:p>Only cell</text:p></table:table-cell>
      </table:table-row>
    </table:table>
  </office:spreadsheet></office:body>
</office:document-content>`,
			want: "Only cell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := argustest.BuildZip(t, argustest.ZipFile{Name: "content.xml", Content: tt.xml})
			got, err := ODS(ctx, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestODS_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	_, err := ODS(ctx, []byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}
