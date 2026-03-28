package extract

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// ODP extracts text from ODP (OpenDocument Presentation) files.
// Parses content.xml with structure-aware handling of presentation pages
// and text frames, preserving slide boundaries.
func ODP(_ context.Context, data []byte) (string, error) {
	r, err := odfContentReader(data)
	if err != nil {
		return "", fmt.Errorf("odp: %w", err)
	}
	defer func() { _ = r.Close() }()
	return parseODFPresentation(r)
}

// parseODFPresentation extracts text from ODF presentation content XML,
// preserving page boundaries and paragraph structure.
func parseODFPresentation(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var pages []string
	var paragraphs []string
	var current strings.Builder
	inPage := false
	inParagraph := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "page":
				inPage = true
			case "p", "h":
				if inPage {
					inParagraph = true
				}
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "p", "h":
				if inParagraph {
					inParagraph = false
					text := strings.TrimSpace(current.String())
					if text != "" {
						paragraphs = append(paragraphs, text)
					}
					current.Reset()
				}
			case "page":
				inPage = false
				if len(paragraphs) > 0 {
					pages = append(pages, strings.Join(paragraphs, "\n"))
					paragraphs = paragraphs[:0]
				}
			}
		case xml.CharData:
			if inParagraph {
				current.Write(t)
			}
		}
	}

	// Flush remaining.
	if text := strings.TrimSpace(current.String()); text != "" {
		paragraphs = append(paragraphs, text)
	}
	if len(paragraphs) > 0 {
		pages = append(pages, strings.Join(paragraphs, "\n"))
	}

	return strings.Join(pages, "\n\n"), nil
}
