//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"

	adminhandlers "github.com/zoobz-io/argus/admin/handlers"
	"github.com/zoobz-io/argus/api/handlers"
	"github.com/zoobz-io/rocco"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// testAPIEngine is the shared API engine, initialized once in TestMain.
var testAPIEngine *rocco.Engine

// testAdminEngine is the shared admin engine, initialized once in TestMain.
var testAdminEngine *rocco.Engine

// testTenantID is the real tenant ID used for API handler tests.
var testTenantID string

// InitEngines creates a test tenant and rocco test engines wired to real handlers.
// Must be called after InitStores.
func InitEngines() error {
	ctx := context.Background()

	// Create a real tenant for API tests (API handlers scope by tenant ID from identity).
	tenant, err := testStores.Tenants.CreateTenant(ctx, "Integration Test Tenant", "integration-test")
	if err != nil {
		return fmt.Errorf("creating test tenant: %w", err)
	}
	testTenantID = tenant.ID

	apiIdentity := rtesting.NewMockIdentity("user-1").WithTenantID(testTenantID)
	testAPIEngine = rtesting.TestEngineWithAuth(func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return apiIdentity, nil
	}).WithHandlers(handlers.All()...)

	adminIdentity := rtesting.NewMockIdentity("admin-1").WithRoles("admin")
	testAdminEngine = rtesting.TestEngineWithAuth(func(_ context.Context, _ *http.Request) (rocco.Identity, error) {
		return adminIdentity, nil
	}).WithHandlers(adminhandlers.All()...)

	return nil
}
