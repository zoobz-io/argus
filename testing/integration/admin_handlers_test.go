//go:build integration

package integration

import (
	"context"
	"testing"

	adminwire "github.com/zoobz-io/argus/admin/wire"
	"github.com/zoobz-io/argus/api/wire"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// =============================================================================
// Tenants
// =============================================================================

func TestAdmin_Tenants_FullCRUD(t *testing.T) {
	// Create.
	body := adminwire.AdminTenantCreateRequest{Name: "Admin Full CRUD", Slug: "admin-full-crud"}
	create := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants", body)
	rtesting.AssertStatus(t, create, 201)

	var created adminwire.AdminTenantResponse
	create.DecodeJSON(&created)

	// Get.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/tenants/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	var got adminwire.AdminTenantResponse
	get.DecodeJSON(&got)
	if got.Name != "Admin Full CRUD" {
		t.Errorf("get mismatch: %+v", got)
	}

	// Update.
	updateBody := adminwire.AdminTenantCreateRequest{Name: "Updated CRUD", Slug: "updated-crud"}
	update := rtesting.ServeRequest(testAdminEngine, "PUT", "/tenants/"+created.ID, updateBody)
	rtesting.AssertStatus(t, update, 200)

	// List.
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/tenants", nil)
	rtesting.AssertStatus(t, list, 200)

	// Delete.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/tenants/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/tenants/"+created.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)
}

func TestAdmin_Tenants_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/tenants/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Providers
// =============================================================================

func TestAdmin_Providers_FullCRUD(t *testing.T) {
	tenantBody := adminwire.AdminTenantCreateRequest{Name: "Provider Admin", Slug: "provider-admin-test"}
	tenantCreate := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants", tenantBody)
	rtesting.AssertStatus(t, tenantCreate, 201)

	var tenant adminwire.AdminTenantResponse
	tenantCreate.DecodeJSON(&tenant)

	// Create.
	provBody := wire.ProviderCreateRequest{Type: "s3", Name: "Admin S3", Credentials: `{"key":"val"}`}
	create := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants/"+tenant.ID+"/providers", provBody)
	rtesting.AssertStatus(t, create, 201)

	var created adminwire.AdminProviderResponse
	create.DecodeJSON(&created)

	// Get.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/providers/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	// List all.
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/providers", nil)
	rtesting.AssertStatus(t, list, 200)

	// Delete.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/providers/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/providers/"+created.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)
}

// =============================================================================
// Watched Paths
// =============================================================================

func TestAdmin_WatchedPaths_FullCRUD(t *testing.T) {
	ctx := context.Background()
	s := Stores(t)

	tenant, _ := s.Tenants.CreateTenant(ctx, "WP Admin", "wp-admin-test")
	provider, _ := s.Providers.CreateProvider(ctx, tenant.ID, "s3", "WP S3", `{}`)
	wp, _ := s.WatchedPaths.CreateWatchedPath(ctx, tenant.ID, provider.ID, "/admin-wp-test")

	// Get.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/watched-paths/"+wp.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	// List all.
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/watched-paths", nil)
	rtesting.AssertStatus(t, list, 200)

	// Delete.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/watched-paths/"+wp.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/watched-paths/"+wp.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)
}

// =============================================================================
// Topics
// =============================================================================

func TestAdmin_Topics_FullCRUD(t *testing.T) {
	tenantBody := adminwire.AdminTenantCreateRequest{Name: "Topic Admin", Slug: "topic-admin-test"}
	tenantCreate := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants", tenantBody)
	rtesting.AssertStatus(t, tenantCreate, 201)

	var tenant adminwire.AdminTenantResponse
	tenantCreate.DecodeJSON(&tenant)

	// Create.
	topicBody := wire.TopicCreateRequest{Name: "Admin Topic", Description: "Admin topic test"}
	create := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants/"+tenant.ID+"/topics", topicBody)
	rtesting.AssertStatus(t, create, 201)

	var created adminwire.AdminTopicResponse
	create.DecodeJSON(&created)

	// Get.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/topics/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	// List.
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/tenants/"+tenant.ID+"/topics", nil)
	rtesting.AssertStatus(t, list, 200)

	// Delete.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/topics/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/topics/"+created.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)
}

// =============================================================================
// Tags
// =============================================================================

func TestAdmin_Tags_FullCRUD(t *testing.T) {
	tenantBody := adminwire.AdminTenantCreateRequest{Name: "Tag Admin", Slug: "tag-admin-test"}
	tenantCreate := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants", tenantBody)
	rtesting.AssertStatus(t, tenantCreate, 201)

	var tenant adminwire.AdminTenantResponse
	tenantCreate.DecodeJSON(&tenant)

	// Create.
	tagBody := wire.TagCreateRequest{Name: "admin-tag", Description: "Admin tag test"}
	create := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants/"+tenant.ID+"/tags", tagBody)
	rtesting.AssertStatus(t, create, 201)

	var created adminwire.AdminTagResponse
	create.DecodeJSON(&created)

	// Get.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/tags/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	// List.
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/tenants/"+tenant.ID+"/tags", nil)
	rtesting.AssertStatus(t, list, 200)

	// Delete.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/tags/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/tags/"+created.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)
}

// =============================================================================
// Documents & Versions
// =============================================================================

func TestAdmin_Documents_AfterIngest(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	doc, version := ingestTestDocument(t, s, ctx)

	// List all documents.
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/documents", nil)
	rtesting.AssertStatus(t, list, 200)

	// Get document.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/documents/"+doc.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	// List all versions.
	listVersions := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions", nil)
	rtesting.AssertStatus(t, listVersions, 200)

	// Get version.
	getVersion := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions/"+version.ID, nil)
	rtesting.AssertStatus(t, getVersion, 200)
}

func TestAdmin_Documents_Delete(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	doc, version := ingestTestDocument(t, s, ctx)

	// Delete version.
	delVersion := rtesting.ServeRequest(testAdminEngine, "DELETE", "/document-versions/"+version.ID, nil)
	rtesting.AssertStatus(t, delVersion, 204)

	getVersion := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions/"+version.ID, nil)
	rtesting.AssertStatus(t, getVersion, 404)

	// Delete document.
	delDoc := rtesting.ServeRequest(testAdminEngine, "DELETE", "/documents/"+doc.ID, nil)
	rtesting.AssertStatus(t, delDoc, 204)

	getDoc := rtesting.ServeRequest(testAdminEngine, "GET", "/documents/"+doc.ID, nil)
	rtesting.AssertStatus(t, getDoc, 404)
}

func TestAdmin_Documents_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/documents/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

func TestAdmin_DocumentVersions_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}
