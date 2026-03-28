package extract

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"
)

// DOCX extracts text from DOCX (Office Open XML) documents.
// Parses word/document.xml with structure-aware handling of paragraphs,
// text runs, tables, and line breaks.
func DOCX(_ context.Context, data []byte) (string, error) {
	r, err := safeZIPReader(data)
	if err != nil {
		return "", fmt.Errorf("opening docx archive: %w", err)
	}

	for _, f := range r.File {
		if f.Name == "word/document.xml" {
			rc, err := safeOpen(f)
			if err != nil {
				return "", fmt.Errorf("opening document.xml: %w", err)
			}
			defer func() { _ = rc.Close() }()
			return parseWordXML(rc)
		}
	}
	return "", fmt.Errorf("word/document.xml not found in archive")
}

// parseWordXML extracts text from WordprocessingML, preserving paragraph
// and table structure.
func parseWordXML(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var paragraphs []string
	var current strings.Builder
	inText := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "t":
				inText = true
			case xmlElemTab:
				current.WriteByte('\t')
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "t":
				inText = false
			case "p":
				text := strings.TrimSpace(current.String())
				if text != "" {
					paragraphs = append(paragraphs, text)
				}
				current.Reset()
			case "br":
				current.WriteByte('\n')
			}
		case xml.CharData:
			if inText {
				current.Write(t)
			}
		}
	}

	// Flush any remaining content.
	if text := strings.TrimSpace(current.String()); text != "" {
		paragraphs = append(paragraphs, text)
	}

	return strings.Join(paragraphs, "\n"), nil
}
