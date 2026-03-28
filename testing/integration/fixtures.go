//go:build integration

package integration

import (
	"archive/zip"
	"bytes"
	"encoding/csv"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/xuri/excelize/v2"
)

// testFixture represents a document fixture for pipeline testing.
type testFixture struct {
	Name     string
	MimeType string
	Content  string // Expected extracted text (substring match)
	Data     []byte
}

// =============================================================================
// Plaintext (text/plain)
// =============================================================================

func buildPlaintextFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		{
			Name: "minimal.txt", MimeType: "text/plain",
			Content: "hello world",
			Data:    []byte("hello world"),
		},
		{
			Name: "sentence.txt", MimeType: "text/plain",
			Content: "quarterly earnings exceeded",
			Data:    []byte("The quarterly earnings exceeded expectations with a 20% increase in revenue."),
		},
		{
			Name: "multiline.txt", MimeType: "text/plain",
			Content: "budget allocation",
			Data:    []byte("Annual Report\n\nThe budget allocation for Q3 was reviewed.\nAll departments met their targets.\nNext steps include finalizing the audit."),
		},
		{
			Name: "unicode.txt", MimeType: "text/plain",
			Content: "international offices",
			Data:    []byte("Report on international offices:\n- Tokyo (東京): Revenue ¥2.3B\n- München: Revenue €1.8M\n- São Paulo: Revenue R$5.2M"),
		},
		{
			Name: "long.txt", MimeType: "text/plain",
			Content: "comprehensive analysis",
			Data:    []byte("This comprehensive analysis covers the full fiscal year performance across all business units. " + strings.Repeat("The detailed metrics show consistent growth in key areas. ", 20) + "In conclusion, the trajectory remains positive."),
		},
	}
}

// =============================================================================
// Markdown (text/markdown)
// =============================================================================

func buildMarkdownFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		{
			Name: "heading-only.md", MimeType: "text/markdown",
			Content: "Project Alpha",
			Data:    []byte("# Project Alpha"),
		},
		{
			Name: "readme.md", MimeType: "text/markdown",
			Content: "infrastructure migration",
			Data:    []byte("# Project Overview\n\nThis document provides details on the infrastructure migration plan.\n\n## Timeline\n\nQ1 2026 through Q3 2026."),
		},
		{
			Name: "notes.md", MimeType: "text/markdown",
			Content: "action items assigned",
			Data:    []byte("## Sprint Review\n\n- Discussed timeline\n- Reviewed sprint velocity\n- Action items assigned to leads\n\n### Decisions\n\n1. Proceed with Phase 2\n2. Defer refactoring to Q4"),
		},
		{
			Name: "table.md", MimeType: "text/markdown",
			Content: "deployment frequency",
			Data:    []byte("# Metrics Dashboard\n\n| Metric | Value |\n|--------|-------|\n| Deployment frequency | 4/week |\n| Lead time | 2 days |\n| MTTR | 1 hour |"),
		},
		{
			Name: "code-blocks.md", MimeType: "text/markdown",
			Content: "configuration reference",
			Data:    []byte("# Configuration Reference\n\nSet the following environment variables:\n\n```bash\nexport APP_PORT=8080\nexport DB_HOST=localhost\n```\n\nSee the configuration reference for details."),
		},
	}
}

// =============================================================================
// CSV (text/csv)
// =============================================================================

func buildCSVFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildCSVFixture(t, "simple.csv", [][]string{
			{"Name", "Value"},
			{"Revenue", "1000000"},
		}, "Revenue"),
		buildCSVFixture(t, "budget.csv", [][]string{
			{"Department", "Budget", "Spent"},
			{"Engineering", "500000", "480000"},
			{"Marketing", "200000", "195000"},
		}, "Engineering"),
		buildCSVFixture(t, "employees.csv", [][]string{
			{"ID", "Name", "Department", "Role"},
			{"1", "Alice Chen", "Engineering", "Senior Developer"},
			{"2", "Bob Smith", "Marketing", "Director"},
			{"3", "Carol Jones", "Finance", "Controller"},
		}, "Alice Chen"),
		buildCSVFixture(t, "timeseries.csv", [][]string{
			{"Date", "Metric", "Value"},
			{"2026-01-01", "CPU Usage", "45.2"},
			{"2026-01-02", "CPU Usage", "52.1"},
			{"2026-01-03", "CPU Usage", "38.7"},
			{"2026-01-04", "CPU Usage", "61.3"},
			{"2026-01-05", "CPU Usage", "44.9"},
		}, "CPU Usage"),
		buildCSVFixture(t, "wide.csv", [][]string{
			{"SKU", "Product", "Category", "Price", "Stock", "Warehouse", "Supplier"},
			{"A001", "Laptop Pro", "Electronics", "1299.99", "150", "West", "TechCorp"},
			{"B002", "Office Chair", "Furniture", "449.00", "300", "East", "FurnishCo"},
		}, "Laptop Pro"),
	}
}

func buildCSVFixture(t *testing.T, name string, rows [][]string, expect string) testFixture {
	t.Helper()
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	for _, row := range rows {
		w.Write(row)
	}
	w.Flush()
	return testFixture{Name: name, MimeType: "text/csv", Content: expect, Data: buf.Bytes()}
}

// =============================================================================
// RTF (application/rtf)
// =============================================================================

func buildRTFFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		{
			Name: "minimal.rtf", MimeType: "application/rtf",
			Content: "hello world",
			Data:    []byte(`{\rtf1 hello world}`),
		},
		{
			Name: "memo.rtf", MimeType: "application/rtf",
			Content: "compliance review",
			Data:    []byte(`{\rtf1\ansi{\fonttbl{\f0 Times New Roman;}}{\colortbl;\red0\green0\blue0;}\f0\fs24 This memo covers the compliance review scheduled for next quarter.\par Action items are pending approval.}`),
		},
		{
			Name: "formatted.rtf", MimeType: "application/rtf",
			Content: "strategic objectives",
			Data:    []byte(`{\rtf1\ansi\f0\fs24 \b Strategic Objectives\b0\par\par 1. Increase market share by 15%\par 2. Launch three new products\par 3. Expand into European markets\par\par These strategic objectives require cross-functional alignment.}`),
		},
		{
			Name: "unicode.rtf", MimeType: "application/rtf",
			Content: "international",
			Data:    []byte(`{\rtf1\ansi\f0\fs24 International report: caf\u233? culture and \u252?ber expansion.}`),
		},
		{
			Name: "nested.rtf", MimeType: "application/rtf",
			Content: "quarterly review",
			Data:    []byte(`{\rtf1\ansi{\fonttbl{\f0 Arial;}}{\stylesheet{\s0 Normal;}}{\info{\author Test Corp}}\f0\fs22 The quarterly review highlighted {\b strong performance} in {\i all divisions}.\par Revenue: \tab $2.3M\par Profit: \tab $450K}`),
		},
	}
}

// =============================================================================
// DOCX (Office Open XML Word)
// =============================================================================

func buildDOCXFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildDOCXFixture(t, "minimal.docx",
			"<w:p><w:r><w:t>Hello from DOCX</w:t></w:r></w:p>",
			"Hello from DOCX"),
		buildDOCXFixture(t, "report.docx",
			"<w:p><w:r><w:t>Annual performance report summary with key metrics.</w:t></w:r></w:p>"+
				"<w:p><w:r><w:t>Revenue increased by 22% year over year.</w:t></w:r></w:p>",
			"performance report"),
		buildDOCXFixture(t, "policy.docx",
			"<w:p><w:r><w:t>Data Retention Policy</w:t></w:r></w:p>"+
				"<w:p><w:r><w:t>All records must be retained for a minimum of 7 years per regulatory requirements.</w:t></w:r></w:p>"+
				"<w:p><w:r><w:t>Exceptions require written approval from the compliance officer.</w:t></w:r></w:p>",
			"retention policy"),
		buildDOCXFixture(t, "tabbed.docx",
			"<w:p><w:r><w:t>Item</w:t><w:tab/><w:t>Amount</w:t></w:r></w:p>"+
				"<w:p><w:r><w:t>Servers</w:t><w:tab/><w:t>$150,000</w:t></w:r></w:p>"+
				"<w:p><w:r><w:t>Licenses</w:t><w:tab/><w:t>$75,000</w:t></w:r></w:p>",
			"Servers"),
		buildDOCXFixture(t, "table.docx",
			"<w:tbl><w:tr>"+
				"<w:tc><w:p><w:r><w:t>Department</w:t></w:r></w:p></w:tc>"+
				"<w:tc><w:p><w:r><w:t>Headcount</w:t></w:r></w:p></w:tc>"+
				"</w:tr><w:tr>"+
				"<w:tc><w:p><w:r><w:t>Engineering</w:t></w:r></w:p></w:tc>"+
				"<w:tc><w:p><w:r><w:t>85</w:t></w:r></w:p></w:tc>"+
				"</w:tr></w:tbl>",
			"Headcount"),
	}
}

func buildDOCXFixture(t *testing.T, name, bodyXML, expect string) testFixture {
	t.Helper()
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>%s</w:body>
</w:document>`, bodyXML)
	return testFixture{
		Name: name, MimeType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		Content: expect, Data: buildZipBytes(t, "word/document.xml", xml),
	}
}

// =============================================================================
// XLSX (Office Open XML Spreadsheet)
// =============================================================================

func buildXLSXFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildXLSXFixture(t, "minimal.xlsx", map[string][][]string{
			"Sheet1": {{"Value"}, {"42"}},
		}, "42"),
		buildXLSXFixture(t, "sales.xlsx", map[string][][]string{
			"Sheet1": {
				{"Region", "Q1 Sales", "Q2 Sales"},
				{"North", "150000", "175000"},
				{"South", "120000", "140000"},
			},
		}, "North"),
		buildXLSXFixture(t, "inventory.xlsx", map[string][][]string{
			"Sheet1": {
				{"Product", "Stock", "Reorder Level", "Supplier"},
				{"Widget A", "500", "100", "Acme Corp"},
				{"Widget B", "250", "50", "Global Parts"},
				{"Widget C", "1000", "200", "Acme Corp"},
			},
		}, "Widget"),
		buildXLSXFixture(t, "multisheet.xlsx", map[string][][]string{
			"Revenue": {
				{"Month", "Amount"},
				{"January", "100000"},
				{"February", "120000"},
			},
			"Expenses": {
				{"Category", "Amount"},
				{"Payroll", "80000"},
				{"Infrastructure", "15000"},
			},
		}, "Payroll"),
		buildXLSXFixture(t, "wide.xlsx", map[string][][]string{
			"Sheet1": {
				{"ID", "Name", "Department", "Location", "Salary", "Start Date", "Status"},
				{"E001", "Alice Chen", "Engineering", "San Francisco", "185000", "2022-03-15", "Active"},
				{"E002", "Bob Smith", "Marketing", "New York", "145000", "2021-08-01", "Active"},
				{"E003", "Carol Jones", "Finance", "Chicago", "165000", "2023-01-10", "Active"},
			},
		}, "Alice Chen"),
	}
}

func buildXLSXFixture(t *testing.T, name string, sheets map[string][][]string, expect string) testFixture {
	t.Helper()
	f := excelize.NewFile()
	first := true
	for sheetName, rows := range sheets {
		if first {
			f.SetSheetName("Sheet1", sheetName)
			first = false
		} else {
			f.NewSheet(sheetName)
		}
		for i, row := range rows {
			for j, cell := range row {
				ref, _ := excelize.CoordinatesToCellName(j+1, i+1)
				f.SetCellValue(sheetName, ref, cell)
			}
		}
	}
	var buf bytes.Buffer
	if err := f.Write(&buf); err != nil {
		t.Fatalf("writing xlsx %s: %v", name, err)
	}
	return testFixture{
		Name: name, MimeType: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		Content: expect, Data: buf.Bytes(),
	}
}

// =============================================================================
// PPTX (Office Open XML Presentation)
// =============================================================================

func buildPPTXFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildPPTXFixture(t, "single-slide.pptx", []string{
			"Welcome to the presentation",
		}, "Welcome"),
		buildPPTXFixture(t, "pitch.pptx", []string{
			"Strategic growth plan for fiscal year 2027",
			"Revenue targets and expansion roadmap",
		}, "strategic growth"),
		buildPPTXFixture(t, "training.pptx", []string{
			"Security Awareness Training",
			"Phishing attacks account for 90% of breaches",
			"Always verify sender identity before clicking links",
		}, "Security Awareness"),
		buildPPTXFixture(t, "quarterly.pptx", []string{
			"Q3 2026 Business Review",
			"Revenue: $4.2M (up 18% YoY)",
			"Customer acquisition cost decreased by 12%",
			"Churn rate stable at 2.1%",
		}, "Business Review"),
		buildPPTXFixture(t, "long-deck.pptx", []string{
			"Annual Technology Strategy",
			"Cloud Migration Progress: 78% complete",
			"Infrastructure costs reduced by $2.1M annually",
			"Security posture improved with zero-trust architecture",
			"Next steps: complete remaining on-premise workload migration",
		}, "Technology Strategy"),
	}
}

func buildPPTXFixture(t *testing.T, name string, slides []string, expect string) testFixture {
	t.Helper()
	var files []zipEntry
	for i, text := range slides {
		xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<p:sld xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main"
       xmlns:p="http://schemas.openxmlformats.org/presentationml/2006/main">
  <p:cSld><p:spTree><p:sp><p:txBody>
    <a:p><a:r><a:t>%s</a:t></a:r></a:p>
  </p:txBody></p:sp></p:spTree></p:cSld>
</p:sld>`, text)
		files = append(files, zipEntry{name: fmt.Sprintf("ppt/slides/slide%d.xml", i+1), content: xml})
	}
	return testFixture{
		Name: name, MimeType: "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		Content: expect, Data: buildZipBytesMulti(t, files),
	}
}

// =============================================================================
// ODT (OpenDocument Text)
// =============================================================================

func buildODTFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildODFFixture(t, "minimal.odt", "application/vnd.oasis.opendocument.text",
			"<text:p>Simple ODT document</text:p>", "Simple ODT"),
		buildODFFixture(t, "letter.odt", "application/vnd.oasis.opendocument.text",
			"<text:p>Dear stakeholder, the annual review summary is attached for your consideration.</text:p>",
			"annual review"),
		buildODFFixture(t, "multi-para.odt", "application/vnd.oasis.opendocument.text",
			"<text:h>Executive Summary</text:h><text:p>Operations performed well in Q3.</text:p><text:p>Revenue targets were met across all regions.</text:p>",
			"Executive Summary"),
		buildODFFixture(t, "formatted.odt", "application/vnd.oasis.opendocument.text",
			"<text:h>Risk Assessment</text:h><text:p>Cybersecurity risks remain the top priority.</text:p><text:p>Supply chain disruptions are a secondary concern.</text:p><text:p>Mitigation strategies are detailed in Appendix B.</text:p>",
			"Risk Assessment"),
		buildODFFixture(t, "tabs-breaks.odt", "application/vnd.oasis.opendocument.text",
			"<text:p>Item<text:tab/>Cost</text:p><text:p>Hardware<text:tab/>$50,000</text:p><text:p>Software<text:tab/>$30,000</text:p>",
			"Hardware"),
	}
}

// =============================================================================
// ODS (OpenDocument Spreadsheet)
// =============================================================================

func buildODSFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildODFFixture(t, "minimal.ods", "application/vnd.oasis.opendocument.spreadsheet",
			"<table:table><table:table-row><table:table-cell><text:p>Cell</text:p></table:table-cell></table:table-row></table:table>",
			"Cell"),
		buildODFFixture(t, "timesheet.ods", "application/vnd.oasis.opendocument.spreadsheet",
			`<table:table><table:table-row><table:table-cell><text:p>Employee</text:p></table:table-cell><table:table-cell><text:p>Hours</text:p></table:table-cell></table:table-row>
			<table:table-row><table:table-cell><text:p>Alice</text:p></table:table-cell><table:table-cell><text:p>40</text:p></table:table-cell></table:table-row></table:table>`,
			"Employee"),
		buildODFFixture(t, "expenses.ods", "application/vnd.oasis.opendocument.spreadsheet",
			`<table:table><table:table-row><table:table-cell><text:p>Category</text:p></table:table-cell><table:table-cell><text:p>Amount</text:p></table:table-cell></table:table-row>
			<table:table-row><table:table-cell><text:p>Travel</text:p></table:table-cell><table:table-cell><text:p>12500</text:p></table:table-cell></table:table-row>
			<table:table-row><table:table-cell><text:p>Equipment</text:p></table:table-cell><table:table-cell><text:p>45000</text:p></table:table-cell></table:table-row></table:table>`,
			"Equipment"),
		buildODFFixture(t, "multi-row.ods", "application/vnd.oasis.opendocument.spreadsheet",
			`<table:table><table:table-row><table:table-cell><text:p>Server</text:p></table:table-cell><table:table-cell><text:p>Status</text:p></table:table-cell><table:table-cell><text:p>Uptime</text:p></table:table-cell></table:table-row>
			<table:table-row><table:table-cell><text:p>prod-web-01</text:p></table:table-cell><table:table-cell><text:p>Healthy</text:p></table:table-cell><table:table-cell><text:p>99.9%</text:p></table:table-cell></table:table-row>
			<table:table-row><table:table-cell><text:p>prod-db-01</text:p></table:table-cell><table:table-cell><text:p>Healthy</text:p></table:table-cell><table:table-cell><text:p>99.95%</text:p></table:table-cell></table:table-row></table:table>`,
			"prod-web-01"),
		buildODFFixture(t, "metrics.ods", "application/vnd.oasis.opendocument.spreadsheet",
			`<table:table><table:table-row><table:table-cell><text:p>KPI</text:p></table:table-cell><table:table-cell><text:p>Target</text:p></table:table-cell><table:table-cell><text:p>Actual</text:p></table:table-cell></table:table-row>
			<table:table-row><table:table-cell><text:p>Customer Satisfaction</text:p></table:table-cell><table:table-cell><text:p>4.5</text:p></table:table-cell><table:table-cell><text:p>4.7</text:p></table:table-cell></table:table-row></table:table>`,
			"Customer Satisfaction"),
	}
}

// =============================================================================
// ODP (OpenDocument Presentation)
// =============================================================================

func buildODPFixtures(t *testing.T) []testFixture {
	t.Helper()
	return []testFixture{
		buildODFFixture(t, "minimal.odp", "application/vnd.oasis.opendocument.presentation",
			"<draw:page><draw:frame><draw:text-box><text:p>Title slide</text:p></draw:text-box></draw:frame></draw:page>",
			"Title slide"),
		buildODFFixture(t, "training.odp", "application/vnd.oasis.opendocument.presentation",
			"<draw:page><draw:frame><draw:text-box><text:p>Security awareness training materials</text:p></draw:text-box></draw:frame></draw:page>",
			"security awareness"),
		buildODFFixture(t, "two-pages.odp", "application/vnd.oasis.opendocument.presentation",
			"<draw:page><draw:frame><draw:text-box><text:p>Onboarding Overview</text:p></draw:text-box></draw:frame></draw:page>"+
				"<draw:page><draw:frame><draw:text-box><text:p>Company policies and procedures</text:p></draw:text-box></draw:frame></draw:page>",
			"Onboarding"),
		buildODFFixture(t, "product-demo.odp", "application/vnd.oasis.opendocument.presentation",
			"<draw:page><draw:frame><draw:text-box><text:p>Product Demo</text:p><text:p>Featuring the new dashboard analytics</text:p></draw:text-box></draw:frame></draw:page>"+
				"<draw:page><draw:frame><draw:text-box><text:p>Live walkthrough of key features</text:p></draw:text-box></draw:frame></draw:page>",
			"dashboard analytics"),
		buildODFFixture(t, "roadmap.odp", "application/vnd.oasis.opendocument.presentation",
			"<draw:page><draw:frame><draw:text-box><text:p>2027 Product Roadmap</text:p></draw:text-box></draw:frame></draw:page>"+
				"<draw:page><draw:frame><draw:text-box><text:p>Q1: Platform migration</text:p></draw:text-box></draw:frame></draw:page>"+
				"<draw:page><draw:frame><draw:text-box><text:p>Q2: API v3 release</text:p></draw:text-box></draw:frame></draw:page>",
			"Product Roadmap"),
	}
}

func buildODFFixture(t *testing.T, name, mimeType, bodyXML, expect string) testFixture {
	t.Helper()
	xml := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<office:document-content
  xmlns:office="urn:oasis:names:tc:opendocument:xmlns:office:1.0"
  xmlns:text="urn:oasis:names:tc:opendocument:xmlns:text:1.0"
  xmlns:table="urn:oasis:names:tc:opendocument:xmlns:table:1.0"
  xmlns:draw="urn:oasis:names:tc:opendocument:xmlns:drawing:1.0">
  <office:body><office:text>%s</office:text></office:body>
</office:document-content>`, bodyXML)
	return testFixture{
		Name: name, MimeType: mimeType,
		Content: expect, Data: buildZipBytes(t, "content.xml", xml),
	}
}

// =============================================================================
// PDF (application/pdf)
// =============================================================================

func buildPDFFixtures(t *testing.T) []testFixture {
	t.Helper()
	data := readProjectTestdata(t, "sample.pdf")
	return []testFixture{
		// Only one PDF fixture since we can't programmatically build valid PDFs.
		// The sample.pdf is hand-crafted with correct xref offsets.
		{Name: "sample.pdf", MimeType: "application/pdf", Content: "Hello from PDF", Data: data},
	}
}

// =============================================================================
// Images (OCR stubbed)
// =============================================================================

func buildImageFixtures(t *testing.T) []testFixture {
	t.Helper()

	// 1x1 PNG
	var pngBuf bytes.Buffer
	pngImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	pngImg.Set(0, 0, color.White)
	png.Encode(&pngBuf, pngImg)

	// 1x1 JPEG
	var jpegBuf bytes.Buffer
	jpegImg := image.NewRGBA(image.Rect(0, 0, 1, 1))
	jpegImg.Set(0, 0, color.RGBA{R: 255, G: 0, B: 0, A: 255})
	jpeg.Encode(&jpegBuf, jpegImg, nil)

	// Larger PNG (10x10 gradient)
	var gradientBuf bytes.Buffer
	gradientImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for x := 0; x < 10; x++ {
		for y := 0; y < 10; y++ {
			gradientImg.Set(x, y, color.RGBA{R: uint8(x * 25), G: uint8(y * 25), B: 128, A: 255})
		}
	}
	png.Encode(&gradientBuf, gradientImg)

	return []testFixture{
		{Name: "white.png", MimeType: "image/png", Content: "", Data: pngBuf.Bytes()},
		{Name: "red.jpg", MimeType: "image/jpeg", Content: "", Data: jpegBuf.Bytes()},
		{Name: "gradient.png", MimeType: "image/png", Content: "", Data: gradientBuf.Bytes()},
		// BMP/TIFF/GIF/WebP can't be easily built in stdlib — using PNG data
		// with the correct MIME type is sufficient since OCR is stubbed.
		{Name: "scan.bmp", MimeType: "image/bmp", Content: "", Data: pngBuf.Bytes()},
		{Name: "document.tiff", MimeType: "image/tiff", Content: "", Data: pngBuf.Bytes()},
	}
}

// =============================================================================
// Legacy Office (converted via LibreOffice sidecar)
// =============================================================================

// buildLegacyOfficeFixtures returns legacy Office format fixtures.
// These go through the LibreOffice convert sidecar → OOXML extractor pipeline.
// Real DOC/XLS/PPT files are needed since LibreOffice must parse them.
// TODO: check in real legacy files to testing/testdata/ for full coverage.
// For now, return nil — the convert sidecar is wired and ready.
func buildLegacyOfficeFixtures(_ *testing.T) []testFixture {
	return nil
}

// =============================================================================
// Helpers
// =============================================================================

type zipEntry struct {
	name    string
	content string
}

func buildZipBytes(t *testing.T, filename, content string) []byte {
	t.Helper()
	return buildZipBytesMulti(t, []zipEntry{{name: filename, content: content}})
}

func buildZipBytesMulti(t *testing.T, entries []zipEntry) []byte {
	t.Helper()
	var buf bytes.Buffer
	w := zip.NewWriter(&buf)
	for _, e := range entries {
		fw, err := w.Create(e.name)
		if err != nil {
			t.Fatalf("creating zip entry %s: %v", e.name, err)
		}
		fw.Write([]byte(e.content))
	}
	w.Close()
	return buf.Bytes()
}

func readProjectTestdata(t *testing.T, name string) []byte {
	t.Helper()
	data, err := readFileFromProjectRoot(name)
	if err != nil {
		t.Fatalf("reading testdata/%s: %v", name, err)
	}
	return data
}

func readFileFromProjectRoot(name string) ([]byte, error) {
	_, filename, _, _ := runtime.Caller(0)
	path := filepath.Join(filepath.Dir(filename), "..", "testdata", name)
	return os.ReadFile(filepath.Clean(path)) //nolint:gosec
}
