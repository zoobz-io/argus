//go:build testing

package wire

import (
	"testing"
)

func TestHookCreateRequest_Validate_Valid(t *testing.T) {
	req := HookCreateRequest{URL: "https://example.com/webhook"}
	if err := req.Validate(); err != nil {
		t.Errorf("expected valid, got error: %v", err)
	}
}

func TestHookCreateRequest_Validate_MissingURL(t *testing.T) {
	req := HookCreateRequest{}
	if err := req.Validate(); err == nil {
		t.Error("expected validation error for empty url")
	}
}

func TestHookCreateRequest_Clone(t *testing.T) {
	orig := HookCreateRequest{URL: "https://example.com/webhook"}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestHookCreateResponse_Clone(t *testing.T) {
	orig := HookCreateResponse{ID: "h1", URL: "https://example.com/webhook", Secret: "whsec_test", Active: true}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestHookResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := HookResponse{ID: "h1", URL: "https://example.com/webhook", Active: true}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "h1" || !resp.Active {
		t.Errorf("unmasked fields mutated: %+v", resp)
	}
}

func TestHookResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := HookResponse{ID: "h1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestHookResponse_Clone(t *testing.T) {
	orig := HookResponse{ID: "h1", URL: "https://example.com/webhook", Active: true}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestHookListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := HookListResponse{
		Hooks: []HookResponse{{ID: "h1"}},
		Limit: 20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHookListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := HookListResponse{Hooks: []HookResponse{{ID: "h1"}}}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestHookListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := HookListResponse{
		Hooks: []HookResponse{{ID: "h1", URL: "https://example.com/webhook"}},
		Limit: 20,
	}
	cloned := orig.Clone()

	cloned.Hooks[0].URL = "mutated"
	if orig.Hooks[0].URL == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestHookListResponse_Clone_NilHooks(t *testing.T) {
	orig := HookListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Hooks != nil {
		t.Error("expected nil Hooks")
	}
}

func TestDeliveryResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := DeliveryResponse{ID: "dl1", HookID: "h1", StatusCode: 200}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ID != "dl1" || resp.StatusCode != 200 {
		t.Errorf("fields mutated: %+v", resp)
	}
}

func TestDeliveryResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := DeliveryResponse{ID: "dl1"}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeliveryResponse_Clone(t *testing.T) {
	orig := DeliveryResponse{ID: "dl1", HookID: "h1", StatusCode: 200, Attempt: 1}
	cloned := orig.Clone()
	if cloned != orig {
		t.Errorf("clone mismatch: %+v vs %+v", cloned, orig)
	}
}

func TestDeliveryListResponse_OnSend(t *testing.T) {
	ctx := setupRegistry(t)
	resp := DeliveryListResponse{
		Deliveries: []DeliveryResponse{{ID: "dl1"}},
		Limit:      20,
	}
	if err := resp.OnSend(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeliveryListResponse_OnSend_Error(t *testing.T) {
	ctx := setupFailRegistry(t)
	resp := DeliveryListResponse{Deliveries: []DeliveryResponse{{ID: "dl1"}}}
	if err := resp.OnSend(ctx); err == nil {
		t.Fatal("expected error")
	}
}

func TestDeliveryListResponse_Clone_SliceIndependence(t *testing.T) {
	orig := DeliveryListResponse{
		Deliveries: []DeliveryResponse{{ID: "dl1", EventID: "evt1"}},
		Limit:      20,
	}
	cloned := orig.Clone()

	cloned.Deliveries[0].EventID = "mutated"
	if orig.Deliveries[0].EventID == "mutated" {
		t.Error("mutating clone affected original")
	}
}

func TestDeliveryListResponse_Clone_NilDeliveries(t *testing.T) {
	orig := DeliveryListResponse{Limit: 10}
	cloned := orig.Clone()
	if cloned.Deliveries != nil {
		t.Error("expected nil Deliveries")
	}
}
