package extract

import (
	"archive/zip"
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"github.com/zoobz-io/sum"

	intcontracts "github.com/zoobz-io/argus/internal/contracts"
)

// LLM returns an extractor that pre-processes document bytes into raw text
// and sends them through a TextExtractor (LLM) for clean content extraction.
func LLM() Func {
	return func(ctx context.Context, data []byte) (string, error) {
		// Try zip-based extraction (OOXML, ODF).
		raw, err := rawZipText(data)
		if err != nil {
			// Not a zip — treat as text-based format (RTF).
			raw = string(data)
		}
		extractor := sum.MustUse[intcontracts.TextExtractor](ctx)
		return extractor.Extract(ctx, raw)
	}
}

// rawZipText opens a zip archive and extracts text from known content files.
// Filters to content-bearing XML only (no styles, themes, metadata).
// Resolves XLSX shared strings to avoid lossy index-number output.
func rawZipText(data []byte) (string, error) {
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return "", fmt.Errorf("opening zip archive: %w", err)
	}

	// Check for XLSX shared strings first.
	shared, _ := xlsxSharedStrings(r)

	var parts []string
	for _, f := range r.File {
		if !isContentFile(f.Name) {
			continue
		}
		// XLSX worksheets need shared string resolution.
		if shared != nil && strings.HasPrefix(f.Name, "xl/worksheets/sheet") {
			text, err := stripSheetXML(f, shared)
			if err != nil {
				return "", fmt.Errorf("parsing %s: %w", f.Name, err)
			}
			if text != "" {
				parts = append(parts, text)
			}
			continue
		}
		text, err := stripXMLFile(f)
		if err != nil {
			return "", fmt.Errorf("parsing %s: %w", f.Name, err)
		}
		if text != "" {
			parts = append(parts, text)
		}
	}
	if len(parts) == 0 {
		return "", fmt.Errorf("no content files found in archive")
	}
	return strings.Join(parts, "\n"), nil
}

// isContentFile returns true if the zip entry contains document content.
func isContentFile(name string) bool {
	switch {
	case name == "word/document.xml":
		return true
	case strings.HasPrefix(name, "xl/worksheets/sheet") && strings.HasSuffix(name, ".xml"):
		return true
	case strings.HasPrefix(name, "ppt/slides/slide") && strings.HasSuffix(name, ".xml"):
		return true
	case name == "content.xml":
		return true
	default:
		return false
	}
}

// stripXMLFile opens a zip entry and returns only its text content with XML tags removed.
func stripXMLFile(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = rc.Close() }()
	return stripXMLTags(rc)
}

// stripXMLTags uses xml.Decoder to collect only CharData from XML content.
func stripXMLTags(r io.Reader) (string, error) {
	decoder := xml.NewDecoder(r)
	var out strings.Builder
	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		if cd, ok := tok.(xml.CharData); ok {
			text := strings.TrimSpace(string(cd))
			if text != "" {
				if out.Len() > 0 {
					out.WriteByte(' ')
				}
				out.WriteString(text)
			}
		}
	}
	return out.String(), nil
}

// xlsxSharedStrings parses the shared string table from an XLSX archive.
func xlsxSharedStrings(r *zip.Reader) ([]string, error) {
	for _, f := range r.File {
		if f.Name == "xl/sharedStrings.xml" {
			rc, err := f.Open()
			if err != nil {
				return nil, fmt.Errorf("opening shared strings: %w", err)
			}
			defer func() { _ = rc.Close() }()

			var result []string
			decoder := xml.NewDecoder(rc)
			var current bytes.Buffer
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
						current.Reset()
					}
				case xml.EndElement:
					if t.Name.Local == "t" {
						inText = false
						result = append(result, current.String())
					}
				case xml.CharData:
					if inText {
						current.Write(t)
					}
				}
			}
			return result, nil
		}
	}
	return nil, nil
}

// stripSheetXML extracts cell text from an XLSX worksheet,
// resolving shared string references to actual text values.
func stripSheetXML(f *zip.File, shared []string) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = rc.Close() }()

	decoder := xml.NewDecoder(rc)
	var rows []string
	var cells []string
	var cellValue strings.Builder
	inValue := false
	isSharedString := false

	for {
		tok, err := decoder.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "c":
				isSharedString = false
				for _, attr := range t.Attr {
					if attr.Name.Local == "t" && attr.Value == "s" {
						isSharedString = true
					}
				}
			case "v":
				inValue = true
				cellValue.Reset()
			}
		case xml.EndElement:
			switch t.Name.Local {
			case "v":
				inValue = false
				val := cellValue.String()
				if isSharedString {
					idx := 0
					for _, ch := range val {
						idx = idx*10 + int(ch-'0')
					}
					if idx < len(shared) {
						val = shared[idx]
					}
				}
				if val != "" {
					cells = append(cells, val)
				}
			case "row":
				if len(cells) > 0 {
					rows = append(rows, strings.Join(cells, "\t"))
					cells = cells[:0]
				}
			}
		case xml.CharData:
			if inValue {
				cellValue.Write(t)
			}
		}
	}
	return strings.Join(rows, "\n"), nil
}
