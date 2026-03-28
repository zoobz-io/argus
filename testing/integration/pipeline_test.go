//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	rtesting "github.com/zoobz-io/rocco/testing"
)

func TestPipeline_IngestPlaintext(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Use the shared test tenant (same tenant as API engine identity).
	tenantID := testTenantID

	provider, err := s.Providers.CreateProvider(ctx, tenantID, models.ProviderS3, "Pipeline S3", `{}`)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}

	wp, err := s.WatchedPaths.CreateWatchedPath(ctx, tenantID, provider.ID, "/pipeline-test")
	if err != nil {
		t.Fatalf("create watched path: %v", err)
	}

	// Create a topic and tag so the analyzer has vocabulary to match.
	_, err = s.Topics.CreateTopic(ctx, tenantID, "Finance", "Financial documents")
	if err != nil {
		t.Fatalf("create topic: %v", err)
	}
	_, err = s.Tags.CreateTag(ctx, tenantID, "quarterly", "Quarterly reports")
	if err != nil {
		t.Fatalf("create tag: %v", err)
	}

	// Upload a plaintext document to MinIO.
	objectKey := "pipeline-test/" + uuid.New().String() + ".txt"
	content := []byte("This is a quarterly financial report with revenue growth of 15% year over year.")
	if err := testInfra.UploadObject(ctx, objectKey, content); err != nil {
		t.Fatalf("upload to minio: %v", err)
	}

	// Create document record in Postgres.
	doc := &models.Document{
		ID:            uuid.New().String(),
		TenantID:      tenantID,
		ProviderID:    provider.ID,
		WatchedPathID: wp.ID,
		ExternalID:    "ext-" + uuid.New().String()[:8],
		Name:          "quarterly-report.txt",
		MimeType:      "text/plain",
		ObjectKey:     objectKey,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	if err := s.Documents.Set(ctx, doc.ID, doc); err != nil {
		t.Fatalf("create document: %v", err)
	}

	// Create document version.
	version := &models.DocumentVersion{
		ID:            uuid.New().String(),
		DocumentID:    doc.ID,
		TenantID:      tenantID,
		VersionNumber: 1,
		ContentHash:   "abc123",
		CreatedAt:     time.Now(),
	}
	if err := s.DocumentVersions.Set(ctx, version.ID, version); err != nil {
		t.Fatalf("create version: %v", err)
	}

	// Trigger ingestion via HTTP — same as a real API call.
	ingestBody := struct {
		VersionID string `json:"version_id"`
	}{VersionID: version.ID}
	ingest := rtesting.ServeRequest(testAPIEngine, "POST", "/ingest", ingestBody)
	if ingest.StatusCode() != 200 {
		t.Fatalf("ingest failed: status %d, body: %s", ingest.StatusCode(), ingest.BodyString())
	}

	// Force OpenSearch to refresh the index so search results are immediate.
	RefreshOpenSearchIndex(ctx, t, testInfra.OpenSearchAddr)

	// Search for the ingested document.
	searchBody := wire.SearchRequest{
		Keywords: "quarterly financial",
		Page:     wire.Page{Size: 10},
	}
	search := rtesting.ServeRequest(testAPIEngine, "POST", "/search", searchBody)
	if search.StatusCode() != 200 {
		t.Fatalf("search failed: status %d, body: %s", search.StatusCode(), search.BodyString())
	}

	var searchResp wire.SearchResponse
	if err := search.DecodeJSON(&searchResp); err != nil {
		t.Fatalf("decode search: %v", err)
	}

	if searchResp.Total == 0 {
		t.Error("expected at least 1 search result after ingestion")
	}
	if len(searchResp.Hits) == 0 {
		t.Fatal("no hits returned")
	}

	hit := searchResp.Hits[0]
	if hit.DocumentID != doc.ID {
		t.Errorf("expected document ID %q, got %q", doc.ID, hit.DocumentID)
	}
	if hit.Summary == "" {
		t.Error("expected non-empty summary from analyzer")
	}

	t.Logf("pipeline complete: doc=%s version=%s summary=%q", doc.ID, version.ID, hit.Summary)
}
