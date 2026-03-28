//go:build integration

package integration

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/argus/models"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// TestPipeline_AllFormats runs the ingestion pipeline for every supported
// MIME type and verifies documents are indexed in OpenSearch.
func TestPipeline_AllFormats(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Use the shared test tenant.
	tenantID := testTenantID

	provider, err := s.Providers.CreateProvider(ctx, tenantID, models.ProviderS3, "Format Test S3", `{}`)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	wp, err := s.WatchedPaths.CreateWatchedPath(ctx, tenantID, provider.ID, "/format-test")
	if err != nil {
		t.Fatalf("create watched path: %v", err)
	}

	// Collect all fixtures.
	var fixtures []testFixture
	fixtures = append(fixtures, buildPlaintextFixtures(t)...)
	fixtures = append(fixtures, buildMarkdownFixtures(t)...)
	fixtures = append(fixtures, buildCSVFixtures(t)...)
	fixtures = append(fixtures, buildRTFFixtures(t)...)
	fixtures = append(fixtures, buildDOCXFixtures(t)...)
	fixtures = append(fixtures, buildXLSXFixtures(t)...)
	fixtures = append(fixtures, buildPPTXFixtures(t)...)
	fixtures = append(fixtures, buildODTFixtures(t)...)
	fixtures = append(fixtures, buildODSFixtures(t)...)
	fixtures = append(fixtures, buildODPFixtures(t)...)
	fixtures = append(fixtures, buildPDFFixtures(t)...)
	fixtures = append(fixtures, buildImageFixtures(t)...)
	fixtures = append(fixtures, buildLegacyOfficeFixtures(t)...)

	t.Logf("testing %d fixtures across all formats", len(fixtures))

	// Track document IDs for later search verification.
	type ingestedDoc struct {
		fixture    testFixture
		documentID string
	}
	var ingested []ingestedDoc

	for _, fix := range fixtures {
		t.Run(fix.Name, func(t *testing.T) {
			// Upload to MinIO.
			objectKey := "format-test/" + uuid.New().String() + "/" + fix.Name
			if err := testInfra.UploadObject(ctx, objectKey, fix.Data); err != nil {
				t.Fatalf("upload %s: %v", fix.Name, err)
			}

			// Create document in Postgres.
			doc := &models.Document{
				ID:            uuid.New().String(),
				TenantID:      tenantID,
				ProviderID:    provider.ID,
				WatchedPathID: wp.ID,
				ExternalID:    "ext-" + uuid.New().String()[:8],
				Name:          fix.Name,
				MimeType:      fix.MimeType,
				ObjectKey:     objectKey,
				CreatedAt:     time.Now(),
				UpdatedAt:     time.Now(),
			}
			if err := s.Documents.Set(ctx, doc.ID, doc); err != nil {
				t.Fatalf("create document: %v", err)
			}

			// Create version.
			version := &models.DocumentVersion{
				ID:            uuid.New().String(),
				DocumentID:    doc.ID,
				TenantID:      tenantID,
				VersionNumber: 1,
				ContentHash:   uuid.New().String()[:16],
				CreatedAt:     time.Now(),
			}
			if err := s.DocumentVersions.Set(ctx, version.ID, version); err != nil {
				t.Fatalf("create version: %v", err)
			}

			// Trigger ingestion.
			body := struct {
				VersionID string `json:"version_id"`
			}{VersionID: version.ID}
			capture := rtesting.ServeRequest(testAPIEngine, "POST", "/ingest", body)
			if capture.StatusCode() != 200 {
				t.Fatalf("ingest failed: status %d, body: %s", capture.StatusCode(), capture.BodyString())
			}

			ingested = append(ingested, ingestedDoc{fixture: fix, documentID: doc.ID})
			t.Logf("ingested %s (%s) -> doc=%s", fix.Name, fix.MimeType, doc.ID)
		})
	}

	// Refresh OpenSearch and verify all documents are searchable.
	RefreshOpenSearchIndex(ctx, t, testInfra.OpenSearchAddr)

	for _, doc := range ingested {
		if doc.fixture.Content == "" {
			continue // OCR stubs don't produce meaningful searchable text
		}
		t.Run("search/"+doc.fixture.Name, func(t *testing.T) {
			searchBody := struct {
				Keywords string `json:"keywords"`
				Page     struct {
					Size int `json:"size"`
				} `json:"page"`
			}{Keywords: doc.fixture.Content}
			searchBody.Page.Size = 50

			capture := rtesting.ServeRequest(testAPIEngine, "POST", "/search", searchBody)
			if capture.StatusCode() != 200 {
				t.Fatalf("search failed: status %d", capture.StatusCode())
			}

			body := capture.BodyString()
			if !strings.Contains(body, doc.documentID) {
				t.Errorf("document %s (%s) not found in search results for %q\nresponse: %s",
					doc.fixture.Name, doc.fixture.MimeType, doc.fixture.Content, body)
			}
		})
	}
}
