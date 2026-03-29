//go:build integration

package integration

import (
	"testing"

	"github.com/zoobz-io/argus/api/wire"
	rtesting "github.com/zoobz-io/rocco/testing"
)

// =============================================================================
// Hooks
// =============================================================================

func TestAPI_Hooks_CRUD(t *testing.T) {
	// Create.
	body := wire.HookCreateRequest{URL: "https://example.com/webhook"}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/hooks", body)
	rtesting.AssertStatus(t, create, 201)

	var created wire.HookCreateResponse
	if err := create.DecodeJSON(&created); err != nil {
		t.Fatalf("decode create: %v", err)
	}
	if created.ID == "" {
		t.Error("expected non-empty hook ID")
	}
	if created.URL != "https://example.com/webhook" {
		t.Errorf("URL: got %q, want %q", created.URL, "https://example.com/webhook")
	}
	if created.Secret == "" {
		t.Error("expected non-empty secret on create response")
	}
	if !created.Active {
		t.Error("expected hook to be active")
	}

	// List.
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/hooks", nil)
	rtesting.AssertStatus(t, list, 200)

	var listResp wire.HookListResponse
	if err := list.DecodeJSON(&listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if len(listResp.Hooks) == 0 {
		t.Error("expected at least one hook in list")
	}

	found := false
	for _, h := range listResp.Hooks {
		if h.ID == created.ID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created hook %s not found in list", created.ID)
	}

	// Delete.
	del := rtesting.ServeRequest(testAPIEngine, "DELETE", "/hooks/"+created.ID, nil)
	rtesting.AssertStatus(t, del, 204)

	// Verify it is gone from list.
	listAfter := rtesting.ServeRequest(testAPIEngine, "GET", "/hooks", nil)
	rtesting.AssertStatus(t, listAfter, 200)

	var listAfterResp wire.HookListResponse
	if err := listAfter.DecodeJSON(&listAfterResp); err != nil {
		t.Fatalf("decode list after delete: %v", err)
	}
	for _, h := range listAfterResp.Hooks {
		if h.ID == created.ID {
			t.Errorf("hook %s should have been deleted", created.ID)
		}
	}
}

func TestAPI_Hooks_NotFound(t *testing.T) {
	del := rtesting.ServeRequest(testAPIEngine, "DELETE", "/hooks/nonexistent", nil)
	rtesting.AssertStatus(t, del, 404)
}

func TestAPI_Hooks_Deliveries(t *testing.T) {
	// Create a hook first.
	body := wire.HookCreateRequest{URL: "https://example.com/deliveries-test"}
	create := rtesting.ServeRequest(testAPIEngine, "POST", "/hooks", body)
	rtesting.AssertStatus(t, create, 201)

	var created wire.HookCreateResponse
	create.DecodeJSON(&created)

	// List deliveries (should be empty initially).
	list := rtesting.ServeRequest(testAPIEngine, "GET", "/hooks/"+created.ID+"/deliveries", nil)
	rtesting.AssertStatus(t, list, 200)

	var listResp wire.DeliveryListResponse
	if err := list.DecodeJSON(&listResp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(listResp.Deliveries) != 0 {
		t.Errorf("expected 0 deliveries, got %d", len(listResp.Deliveries))
	}

	// Cleanup.
	rtesting.ServeRequest(testAPIEngine, "DELETE", "/hooks/"+created.ID, nil)
}

// =============================================================================
// Admin: Audit Log
// =============================================================================

func TestAdmin_AuditLog(t *testing.T) {
	list := rtesting.ServeRequest(testAdminEngine, "GET", "/audit", nil)
	rtesting.AssertStatus(t, list, 200)

	var resp wire.AuditListResponse
	if err := list.DecodeJSON(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	// The endpoint should return successfully even if there are no entries.
	t.Logf("audit log returned %d entries", len(resp.Entries))
}
