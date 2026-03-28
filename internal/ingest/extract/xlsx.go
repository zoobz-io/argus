package extract

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/xuri/excelize/v2"
)

// XLSX extracts text from XLSX (Office Open XML) spreadsheets using excelize.
// Preserves row/column structure with tab-separated cells and newline-separated rows.
// Each sheet is separated by a double newline.
func XLSX(_ context.Context, data []byte) (string, error) {
	// Pre-validate archive before handing to excelize which handles zip internally.
	if _, err := safeZIPReader(data); err != nil {
		return "", fmt.Errorf("xlsx archive validation: %w", err)
	}

	f, err := excelize.OpenReader(bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("opening xlsx: %w", err)
	}
	defer func() { _ = f.Close() }()

	var sheets []string
	for _, sheet := range f.GetSheetList() {
		rows, err := f.GetRows(sheet)
		if err != nil {
			return "", fmt.Errorf("reading sheet %q: %w", sheet, err)
		}

		var lines []string
		for _, row := range rows {
			// Skip fully empty rows.
			hasContent := false
			for _, cell := range row {
				if cell != "" {
					hasContent = true
					break
				}
			}
			if hasContent {
				lines = append(lines, strings.Join(row, "\t"))
			}
		}

		if len(lines) > 0 {
			sheets = append(sheets, strings.Join(lines, "\n"))
		}
	}

	return strings.Join(sheets, "\n\n"), nil
}
