package extract

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ODT extracts text from ODT (OpenDocument Text) documents.
// Parses content.xml with structure-aware handling of ODF text elements.
func ODT(_ context.Context, data []byte) (string, error) {
	r, err := odfContentReader(data)
	if err != nil {
		return "", fmt.Errorf("odt: %w", err)
	}
	defer func() { _ = r.Close() }()
	return parseODFText(r)
}

// odfContentReader opens the content.xml from an ODF archive with safety checks.
func odfContentReader(data []byte) (io.ReadCloser, error) {
	zr, err := safeZIPReader(data)
	if err != nil {
		return nil, fmt.Errorf("opening archive: %w", err)
	}
	for _, f := range zr.File {
		if f.Name == "content.xml" {
			return safeOpen(f)
		}
	}
	return nil, fmt.Errorf("content.xml not found in archive")
}

// parseODFText extracts text from ODF content XML, preserving paragraphs.
// Handles <text:p>, <text:h>, <text:tab>, and <text:line-break> elements.
func parseODFText(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var paragraphs []string
	var current strings.Builder
	depth := 0

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p", "h":
				depth++
			case xmlElemTab:
				if depth > 0 {
					current.WriteByte('\t')
				}
			case "line-break":
				if depth > 0 {
					current.WriteByte('\n')
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p", "h":
				depth--
				if depth <= 0 {
					text := strings.TrimSpace(current.String())
					if text != "" {
						paragraphs = append(paragraphs, text)
					}
					current.Reset()
					depth = 0
				}
			}
		case xml.CharData:
			if depth > 0 {
				current.Write(t)
			}
		}
	}

	if text := strings.TrimSpace(current.String()); text != "" {
		paragraphs = append(paragraphs, text)
	}

	return strings.Join(paragraphs, "\n"), nil
}
