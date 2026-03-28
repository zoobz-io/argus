//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/zoobz-io/argus/api/wire"
	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/stores"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// =============================================================================
// Topics
// =============================================================================

func TestAPI_Topics_CRUD(t *testing.T) {
	// Create.
	body := wire.TopicCreateRequest{Name: "API CRUD Topic", Description: "Full CRUD test"}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/topics", body)
	rtesting.AssertStatus(t, create, 201)

	var created wire.TopicResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == "" || created.Name != "API CRUD Topic" {
		t.Errorf("create mismatch: %+v", created)
	}

	// Get.
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/topics/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	var got wire.TopicResponse
	if err := get.DecodeJSON(&got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("get mismatch: %+v", got)
	}

	// Update.
	updateBody := wire.TopicCreateRequest{Name: "Updated Topic", Description: "Updated description"}
	update := rtesting.ServeRequest(testAPIEngine, "PUT", "/topics/"+created.ID, updateBody)
	rtesting.AssertStatus(t, update, 200)

	var updated wire.TopicResponse
	if err := update.DecodeJSON(&updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Name != "Updated Topic" {
		t.Errorf("update mismatch: %+v", updated)
	}

	// List.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/topics", nil)
	rtesting.AssertStatus(t, list, 200)

	var listResp wire.TopicListResponse
	if err := list.DecodeJSON(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Topics) == 0 {
		t.Error("expected at least one topic")
	}
}

func TestAPI_Topics_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/topics/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Tags
// =============================================================================

func TestAPI_Tags_CRUD(t *testing.T) {
	body := wire.TagCreateRequest{Name: "api-crud-tag", Description: "Full CRUD test"}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/tags", body)
	rtesting.AssertStatus(t, create, 201)

	var created wire.TagResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Get.
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/tags/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	// Update.
	updateBody := wire.TagCreateRequest{Name: "updated-tag", Description: "Updated tag"}
	update := rtesting.ServeRequest(testAPIEngine, "PUT", "/tags/"+created.ID, updateBody)
	rtesting.AssertStatus(t, update, 200)

	// List.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/tags", nil)
	rtesting.AssertStatus(t, list, 200)

	var listResp wire.TagListResponse
	if err := list.DecodeJSON(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Tags) == 0 {
		t.Error("expected at least one tag")
	}
}

func TestAPI_Tags_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/tags/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Providers
// =============================================================================

func TestAPI_Providers_CRUD(t *testing.T) {
	body := wire.ProviderCreateRequest{Type: "s3", Name: "API CRUD S3", Credentials: `{"key":"val"}`}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/providers", body)
	rtesting.AssertStatus(t, create, 201)

	var created wire.ProviderResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Get.
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/providers/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	var got wire.ProviderResponse
	if err := get.DecodeJSON(&got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.ID != created.ID || got.Name != "API CRUD S3" {
		t.Errorf("get mismatch: %+v", got)
	}

	// Update.
	updateBody := wire.ProviderCreateRequest{Type: "gcs", Name: "Updated GCS", Credentials: `{"new":"creds"}`}
	update := rtesting.ServeRequest(testAPIEngine, "PUT", "/providers/"+created.ID, updateBody)
	rtesting.AssertStatus(t, update, 200)

	var updated wire.ProviderResponse
	if err := update.DecodeJSON(&updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Name != "Updated GCS" {
		t.Errorf("update mismatch: %+v", updated)
	}

	// List.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/providers", nil)
	rtesting.AssertStatus(t, list, 200)
}

func TestAPI_Providers_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/providers/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Watched Paths
// =============================================================================

func TestAPI_WatchedPaths_CRUD(t *testing.T) {
	// Create provider first.
	provBody := wire.ProviderCreateRequest{Type: "s3", Name: "WP CRUD Provider", Credentials: `{}`}
	provCreate := rtesting.ServeRequest(testAPIEngine, "POST", "/providers", provBody)
	rtesting.AssertStatus(t, provCreate, 201)

	var prov wire.ProviderResponse
	provCreate.DecodeJSON(&prov)

	// Create watched path.
	wpBody := wire.WatchedPathCreateRequest{Path: "/crud/test", ProviderID: prov.ID}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/watched-paths", wpBody)
	rtesting.AssertStatus(t, create, 201)

	var created wire.WatchedPathResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode: %v", err)
	}

	// Get.
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/watched-paths/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	var got wire.WatchedPathResponse
	get.DecodeJSON(&got)
	if got.Path != "/crud/test" {
		t.Errorf("get mismatch: %+v", got)
	}

	// Update.
	updateBody := wire.WatchedPathCreateRequest{Path: "/crud/updated"}
	update := rtesting.ServeRequest(testAPIEngine, "PUT", "/watched-paths/"+created.ID, updateBody)
	rtesting.AssertStatus(t, update, 200)

	// List.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/watched-paths", nil)
	rtesting.AssertStatus(t, list, 200)
}

func TestAPI_WatchedPaths_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/watched-paths/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Documents & Versions (require ingested data)
// =============================================================================

func TestAPI_Documents_AfterIngest(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Ingest a document so we have data.
	doc, version := ingestTestDocument(t, s, ctx)

	// List documents.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/documents", nil)
	rtesting.AssertStatus(t, list, 200)

	// Get document.
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/documents/"+doc.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	var gotDoc wire.DocumentResponse
	if err := get.DecodeJSON(&gotDoc); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if gotDoc.ID != doc.ID || gotDoc.Name != doc.Name {
		t.Errorf("document mismatch: %+v", gotDoc)
	}

	// List versions.
	listVersions := rtesting.ServeRequest(testAPIEngine, "GET", "/documents/"+doc.ID+"/versions", nil)
	rtesting.AssertStatus(t, listVersions, 200)

	// Get version.
	getVersion := rtesting.ServeRequest(testAPIEngine, "GET", "/documents/"+doc.ID+"/versions/"+version.ID, nil)
	rtesting.AssertStatus(t, getVersion, 200)

	var gotVersion wire.DocumentVersionResponse
	if err := getVersion.DecodeJSON(&gotVersion); err != nil {
		t.Fatalf("decode version: %v", err)
	}
	if gotVersion.ID != version.ID {
		t.Errorf("version mismatch: %+v", gotVersion)
	}
}

func TestAPI_Documents_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/documents/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

func TestAPI_DocumentVersions_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAPIEngine, "GET", "/documents/any/versions/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Search & Related
// =============================================================================

func TestAPI_Search(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Ingest a document.
	ingestTestDocument(t, s, ctx)
	RefreshOpenSearchIndex(ctx, t, testInfra.OpenSearchAddr)

	searchBody := wire.SearchRequest{
		Keywords: "searchable content",
		Page:     wire.Page{Size: 10},
	}
	search := rtesting.ServeRequest(testAPIEngine, "POST", "/search", searchBody)
	rtesting.AssertStatus(t, search, 200)

	var resp wire.SearchResponse
	if err := search.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// May or may not find results depending on content, but the endpoint should work.
	t.Logf("search returned %d hits", resp.Total)
}

func TestAPI_RelatedDocuments(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	doc, _ := ingestTestDocument(t, s, ctx)
	RefreshOpenSearchIndex(ctx, t, testInfra.OpenSearchAddr)

	searchBody := wire.SearchRequest{Page: wire.Page{Size: 10}}
	related := rtesting.ServeRequest(testAPIEngine, "POST", "/documents/"+doc.ID+"/related", searchBody)
	rtesting.AssertStatus(t, related, 200)
}

func TestAPI_RelatedDocuments_NotFound(t *testing.T) {
	searchBody := wire.SearchRequest{Page: wire.Page{Size: 10}}
	related := rtesting.ServeRequest(testAPIEngine, "POST", "/documents/nonexistent/related", searchBody)
	rtesting.AssertStatus(t, related, 404)
}

// =============================================================================
// Ingest
// =============================================================================

func TestAPI_TriggerIngest(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	_, version := setupIngestableDocument(t, s, ctx)

	body := struct {
		VersionID string `json:"version_id"`
	}{VersionID: version.ID}
	ingest := rtesting.ServeRequest(testAPIEngine, "POST", "/ingest", body)
	rtesting.AssertStatus(t, ingest, 200)
}

// =============================================================================
// Helpers
// =============================================================================

// ingestTestDocument creates a document, uploads content, and runs the pipeline.
// Returns the document and version for further assertions.
func ingestTestDocument(t *testing.T, s *stores.Stores, ctx context.Context) (*models.Document, *models.DocumentVersion) {
	t.Helper()
	doc, version := setupIngestableDocument(t, s, ctx)

	// Trigger ingest.
	body := struct {
		VersionID string `json:"version_id"`
	}{VersionID: version.ID}
	ingest := rtesting.ServeRequest(testAPIEngine, "POST", "/ingest", body)
	if ingest.StatusCode() != 200 {
		t.Fatalf("ingest failed: %d %s", ingest.StatusCode(), ingest.BodyString())
	}

	return doc, version
}

// setupIngestableDocument creates all prerequisite data and uploads content to MinIO.
func setupIngestableDocument(t *testing.T, s *stores.Stores, ctx context.Context) (*models.Document, *models.DocumentVersion) {
	t.Helper()
	tenantID := testTenantID

	provider, err := s.Providers.CreateProvider(ctx, tenantID, models.ProviderS3, "Handler Test S3", `{}`)
	if err != nil {
		t.Fatalf("create provider: %v", err)
	}
	wp, err := s.WatchedPaths.CreateWatchedPath(ctx, tenantID, provider.ID, "/handler-test")
	if err != nil {
		t.Fatalf("create watched path: %v", err)
	}

	objectKey := "handler-test/" + uuid.New().String() + ".txt"
	content := []byte("This is searchable content for handler integration tests.")
	if err := testInfra.UploadObject(ctx, objectKey, content); err != nil {
		t.Fatalf("upload: %v", err)
	}

	doc := &models.Document{
		ID: uuid.New().String(), TenantID: tenantID, ProviderID: provider.ID,
		WatchedPathID: wp.ID, ExternalID: "ext-" + uuid.New().String()[:8],
		Name: "handler-test.txt", MimeType: "text/plain", ObjectKey: objectKey,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	if err := s.Documents.Set(ctx, doc.ID, doc); err != nil {
		t.Fatalf("create doc: %v", err)
	}

	version := &models.DocumentVersion{
		ID: uuid.New().String(), DocumentID: doc.ID, TenantID: tenantID,
		VersionNumber: 1, ContentHash: uuid.New().String()[:16],
		CreatedAt: time.Now(),
	}
	if err := s.DocumentVersions.Set(ctx, version.ID, version); err != nil {
		t.Fatalf("create version: %v", err)
	}

	return doc, version
}
