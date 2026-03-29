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

func TestAdmin_DocumentVersions_DeleteSpecificVersion(t *testing.T) {
	s := Stores(t)
	ctx := context.Background()

	// Ingest to create a document and version.
	doc, version := ingestTestDocument(t, s, ctx)

	// Verify the version exists before deletion.
	getBefore := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions/"+version.ID, nil)
	rtesting.AssertStatus(t, getBefore, 200)

	// Delete only the version.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/document-versions/"+version.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Version should be gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions/"+version.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)

	// Document should still exist after version deletion.
	getDoc := rtesting.ServeRequest(testAdminEngine, "GET", "/documents/"+doc.ID, nil)
	rtesting.AssertStatus(t, getDoc, 200)
}

// =============================================================================
// Users
// =============================================================================

func TestAdmin_Users_List(t *testing.T) {
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/users", nil)
	rtesting.AssertStatus(t, list, 200)

	var resp adminwire.AdminUserListResponse
	if err := list.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// At least the test user created in InitEngines should be present.
	if len(resp.Users) == 0 {
		t.Error("expected at least one user")
	}
}

func TestAdmin_Users_CRUD(t *testing.T) {
	// Create a tenant to own the user.
	tenantBody := adminwire.AdminTenantCreateRequest{Name: "User Admin", Slug: "user-admin-test"}
	tenantCreate := rtesting.ServeRequest(testAdminEngine, "POST", "/tenants", tenantBody)
	rtesting.AssertStatus(t, tenantCreate, 201)

	var tenant adminwire.AdminTenantResponse
	tenantCreate.DecodeJSON(&tenant)

	// Create user.
	userBody := adminwire.AdminUserCreateRequest{
		TenantID:    tenant.ID,
		ExternalID:  "ext-admin-crud-user",
		Email:       "admin-crud@example.com",
		DisplayName: "Admin CRUD User",
		Role:        "viewer",
	}
	create := rtesting.ServeRequest(testAdminEngine, "POST", "/users", userBody)
	rtesting.AssertStatus(t, create, 201)

	var created adminwire.AdminUserResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == "" || created.Email != "admin-crud@example.com" {
		t.Errorf("create mismatch: %+v", created)
	}

	// Get.
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/users/"+created.ID, nil)
	rtesting.AssertStatus(t, get, 200)

	var got adminwire.AdminUserResponse
	if err := get.DecodeJSON(&got); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("get mismatch: %+v", got)
	}

	// Update.
	newEmail := "updated@example.com"
	newName := "Updated User"
	updateBody := adminwire.AdminUserUpdateRequest{
		Email:       &newEmail,
		DisplayName: &newName,
	}
	update := rtesting.ServeRequest(testAdminEngine, "PUT", "/users/"+created.ID, updateBody)
	rtesting.AssertStatus(t, update, 200)

	var updated adminwire.AdminUserResponse
	if err := update.DecodeJSON(&updated); err != nil {
		t.Fatalf("decode update: %v", err)
	}
	if updated.Email != "updated@example.com" || updated.DisplayName != "Updated User" {
		t.Errorf("update mismatch: %+v", updated)
	}

	// Delete.
	del := rtesting.ServeRequest(testAdminEngine, "DELETE", "/users/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify gone.
	getAfter := rtesting.ServeRequest(testAdminEngine, "GET", "/users/"+created.ID, nil)
	rtesting.AssertStatus(t, getAfter, 404)
}

func TestAdmin_Users_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/users/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

// =============================================================================
// Documents & Versions
// =============================================================================

func TestAdmin_Documents_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/documents/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}

func TestAdmin_DocumentVersions_NotFound(t *testing.T) {
	get := rtesting.ServeRequest(testAdminEngine, "GET", "/document-versions/nonexistent", nil)
	rtesting.AssertStatus(t, get, 404)
}
