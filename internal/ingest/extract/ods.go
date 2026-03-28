package extract

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ODS extracts text from ODS (OpenDocument Spreadsheet) files.
// Parses content.xml with structure-aware handling of table rows and cells.
func ODS(_ context.Context, data []byte) (string, error) {
	r, err := odfContentReader(data)
	if err != nil {
		return "", fmt.Errorf("ods: %w", err)
	}
	defer func() { _ = r.Close() }()
	return parseODFTable(r)
}

// parseODFTable extracts text from ODF spreadsheet content XML,
// preserving row/column structure with tab-separated cells and newline-separated rows.
func parseODFTable(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var rows []string
	var cells []string
	var current strings.Builder
	inCell := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "table-cell" {
				inCell = true
				current.Reset()
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "table-cell":
				inCell = false
				cells = append(cells, strings.TrimSpace(current.String()))
			case "table-row":
				// Skip empty rows.
				hasContent := false
				for _, c := range cells {
					if c != "" {
						hasContent = true
						break
					}
				}
				if hasContent {
					rows = append(rows, strings.Join(cells, "\t"))
				}
				cells = cells[:0]
			}
		case xml.CharData:
			if inCell {
				current.Write(t)
			}
		}
	}

	return strings.Join(rows, "\n"), nil
}
