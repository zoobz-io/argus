package extract

import (
	"bytes"
	"context"
	"testing"

	"github.com/xuri/excelize/v2"
)

// buildXLSX creates a minimal XLSX file in memory using excelize.
func buildXLSX(t *testing.T, sheets map[string][][]string) []byte {
	t.Helper()
	f := excelize.NewFile()
	first := true
	for name, rows := range sheets {
		if first {
			// Rename the default sheet.
			if err := f.SetSheetName("Sheet1", name); err != nil {
				t.Fatalf("renaming sheet: %v", err)
			}
			first = false
		} else {
			if _, err := f.NewSheet(name); err != nil {
				t.Fatalf("creating sheet %s: %v", name, err)
			}
		}
		for i, row := range rows {
			for j, cell := range row {
				cellRef, err := excelize.CoordinatesToCellName(j+1, i+1)
				if err != nil {
					t.Fatalf("coordinates: %v", err)
				}
				if err := f.SetCellValue(name, cellRef, cell); err != nil {
					t.Fatalf("setting cell %s: %v", cellRef, err)
				}
			}
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("writing xlsx: %v", err)
	}
	return buf.Bytes()
}

func TestXLSX(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name   string
		sheets map[string][][]string
		want   string
	}{
		{
			name: "single sheet with data",
			sheets: map[string][][]string{
				"Data": {
					{"Name", "Age"},
					{"Alice", "30"},
					{"Bob", "25"},
				},
			},
			want: "Name\tAge\nAlice\t30\nBob\t25",
		},
		{
			name: "empty rows are skipped",
			sheets: map[string][][]string{
				"Data": {
					{"Row 1"},
					{},
					{"Row 3"},
				},
			},
			want: "Row 1\nRow 3",
		},
		{
			name: "single cell",
			sheets: map[string][][]string{
				"Data": {
					{"Only cell"},
				},
			},
			want: "Only cell",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := buildXLSX(t, tt.sheets)
			got, err := XLSX(ctx, data)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestXLSX_InvalidArchive(t *testing.T) {
	ctx := context.Background()
	_, err := XLSX(ctx, []byte("not a zip"))
	if err == nil {
		t.Fatal("expected error for invalid archive")
	}
}
