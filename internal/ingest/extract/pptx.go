package extract

import (
	"archive/zip"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
)

// PPTX extracts text from PPTX (Office Open XML) presentations.
// Parses each slide XML with structure-aware handling of text frames
// and paragraphs, preserving slide boundaries.
func PPTX(_ context.Context, data []byte) (string, error) {
	r, err := safeZIPReader(data)
	if err != nil {
		return "", fmt.Errorf("opening pptx archive: %w", err)
	}

	// Collect slide files in order.
	var slides []*zip.File
	for _, f := range r.File {
		if strings.HasPrefix(f.Name, "ppt/slides/slide") && strings.HasSuffix(f.Name, ".xml") {
			slides = append(slides, f)
		}
	}
	sort.Slice(slides, func(i, j int) bool {
		return slides[i].Name < slides[j].Name
	})

	var parts []string
	for _, f := range slides {
		rc, err := safeOpen(f)
		if err != nil {
			return "", fmt.Errorf("opening %s: %w", f.Name, err)
		}
		text, _ := parseSlideXML(rc)
		_ = rc.Close()
		if text != "" {
			parts = append(parts, text)
		}
	}

	return strings.Join(parts, "\n\n"), nil
}

// parseSlideXML extracts text from a PresentationML slide, preserving
// paragraph structure within text frames.
func parseSlideXML(r io.Reader) (string, error) {
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
			if t.Name.Local == "t" {
				inText = true
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
			}
		case xml.CharData:
			if inText {
				current.Write(t)
			}
		}
	}

	if text := strings.TrimSpace(current.String()); text != "" {
		paragraphs = append(paragraphs, text)
	}

	return strings.Join(paragraphs, "\n"), nil
}
