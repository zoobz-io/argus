package ingest

import (
	"context"
	"fmt"

	"github.com/zoobz-io/capitan"
	"github.com/zoobz-io/pipz"
	"github.com/zoobz-io/sum"

	"github.com/zoobz-io/argus/events"
	intcontracts "github.com/zoobz-io/argus/internal/contracts"
	"github.com/zoobz-io/argus/internal/ingest/extract"
)

// extractRoute wraps an extract.Func into a pipz processor that sets dc.Content.
func extractRoute(id pipz.Identity, fn extract.Func) pipz.Chainable[*DocumentContext] {
	return pipz.Apply(id, func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
		content, err := fn(ctx, dc.RawBytes)
		if err != nil {
			return dc, err
		}
		dc.Content = content
		return dc, nil
	})
}

func newExtractStage() pipz.Chainable[*DocumentContext] {
	// Fetch raw bytes from storage.
	fetch := pipz.Apply(
		FetchID,
		func(ctx context.Context, dc *DocumentContext) (*DocumentContext, error) {
			versions := sum.MustUse[intcontracts.IngestVersions](ctx)
			data, err := versions.GetVersionContent(ctx, dc.Document.ObjectKey)
			if err != nil {
				return dc, fmt.Errorf("fetching object %s: %w", dc.Document.ObjectKey, err)
			}
			dc.RawBytes = data
			return dc, nil
		},
	)

	// Route by MIME type to format-specific extractors.
	router := pipz.NewSwitch(
		ExtractRouterID,
		func(_ context.Context, dc *DocumentContext) string {
			return dc.Document.MimeType
		},
	)

	// Images → OCR.
	for _, mime := range []string{
		extract.MimePNG, extract.MimeJPEG, extract.MimeTIFF,
		extract.MimeBMP, extract.MimeGIF, extract.MimeWebP,
	} {
		router.AddRoute(mime, extractRoute(
			pipz.NewIdentity("extract-"+mime, "OCR extraction for "+mime),
			extract.OCR(mime),
		))
	}

	// PDF → text extraction with OCR fallback.
	router.AddRoute(extract.MimePDF, extractRoute(
		pipz.NewIdentity("extract-pdf", "PDF text extraction with OCR fallback"),
		extract.PDF,
	))

	// Legacy Office → Convert (LibreOffice) → OOXML extractor.
	for _, mime := range []string{extract.MimeDOC, extract.MimeXLS, extract.MimePPT} {
		router.AddRoute(mime, extractRoute(
			pipz.NewIdentity("extract-"+mime, "Convert and extract for "+mime),
			extract.Convert(mime),
		))
	}

	// Plaintext formats → direct read.
	for _, mime := range []string{extract.MimePlain, extract.MimeMarkdown, extract.MimeCSV} {
		router.AddRoute(mime, extractRoute(
			pipz.NewIdentity("extract-"+mime, "Plaintext extraction for "+mime),
			extract.Plaintext,
		))
	}

	// Office Open XML → format-specific parsers.
	router.AddRoute(extract.MimeDOCX, extractRoute(
		pipz.NewIdentity("extract-docx", "DOCX text extraction"),
		extract.DOCX,
	))
	router.AddRoute(extract.MimeXLSX, extractRoute(
		pipz.NewIdentity("extract-xlsx", "XLSX text extraction"),
		extract.XLSX,
	))
	router.AddRoute(extract.MimePPTX, extractRoute(
		pipz.NewIdentity("extract-pptx", "PPTX text extraction"),
		extract.PPTX,
	))

	// OpenDocument → format-specific parsers.
	router.AddRoute(extract.MimeODT, extractRoute(
		pipz.NewIdentity("extract-odt", "ODT text extraction"),
		extract.ODT,
	))
	router.AddRoute(extract.MimeODS, extractRoute(
		pipz.NewIdentity("extract-ods", "ODS text extraction"),
		extract.ODS,
	))
	router.AddRoute(extract.MimeODP, extractRoute(
		pipz.NewIdentity("extract-odp", "ODP text extraction"),
		extract.ODP,
	))

	// RTF → control code stripping.
	router.AddRoute(extract.MimeRTF, extractRoute(
		pipz.NewIdentity("extract-rtf", "RTF text extraction"),
		extract.RTF,
	))

	// Guard: reject unrecognized MIME types.
	// pipz.Switch passes through unchanged on unmatched keys — catch that here.
	// Uses HasRoute rather than checking Content to avoid false positives on
	// extractors that legitimately return empty content (blank scanned pages, etc.).
	guard := pipz.Apply(
		pipz.NewIdentity("extract-guard", "Reject unsupported MIME types"),
		func(_ context.Context, dc *DocumentContext) (*DocumentContext, error) {
			if !router.HasRoute(dc.Document.MimeType) {
				return dc, fmt.Errorf("unsupported MIME type: %s", dc.Document.MimeType)
			}
			return dc, nil
		},
	)

	// Emit signal after extraction.
	signal := pipz.Effect(
		ExtractSignalID,
		func(ctx context.Context, dc *DocumentContext) error {
			capitan.Info(ctx, events.IngestExtracted,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
				events.IngestJobIDKey.Field(dc.Job.ID),
				events.IngestMimeTypeKey.Field(dc.Document.MimeType),
			)
			return nil
		},
	)

	return pipz.NewSequence(ExtractID, fetch, router, guard, signal)
}
