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
			data, err := versions.GetVersionContent(ctx, dc.Version.ObjectKey)
			if err != nil {
				return dc, fmt.Errorf("fetching object %s: %w", dc.Version.ObjectKey, err)
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

	// PDF → OCR.
	router.AddRoute(extract.MimePDF, extractRoute(
		pipz.NewIdentity("extract-pdf", "PDF text extraction"),
		extract.PDF,
	))

	// Legacy Office → OCR.
	for _, mime := range []string{extract.MimeDOC, extract.MimeXLS, extract.MimePPT} {
		router.AddRoute(mime, extractRoute(
			pipz.NewIdentity("extract-"+mime, "Legacy extraction for "+mime),
			extract.Legacy(mime),
		))
	}

	// Plaintext formats → direct read.
	for _, mime := range []string{extract.MimePlain, extract.MimeMarkdown, extract.MimeCSV} {
		router.AddRoute(mime, extractRoute(
			pipz.NewIdentity("extract-"+mime, "Plaintext extraction for "+mime),
			extract.Plaintext,
		))
	}

	// Document formats → LLM extraction (pre-process + transform synapse).
	llmExtractor := extract.LLM()
	for _, mime := range []string{
		extract.MimeRTF,
		extract.MimeDOCX, extract.MimeXLSX, extract.MimePPTX,
		extract.MimeODT, extract.MimeODS, extract.MimeODP,
	} {
		router.AddRoute(mime, extractRoute(
			pipz.NewIdentity("extract-"+mime, "LLM extraction for "+mime),
			llmExtractor,
		))
	}

	// Emit signal after extraction.
	signal := pipz.Effect(
		ExtractSignalID,
		func(ctx context.Context, dc *DocumentContext) error {
			capitan.Info(ctx, events.IngestExtracted,
				events.IngestVersionIDKey.Field(dc.Version.ID),
				events.IngestDocumentIDKey.Field(dc.Version.DocumentID),
				events.IngestMimeTypeKey.Field(dc.Document.MimeType),
			)
			return nil
		},
	)

	return pipz.NewSequence(ExtractID, fetch, router, signal)
}
