//go:build testing

package transformers

import (
	"testing"

	"github.com/zoobz-io/argus/models"
	argustest "github.com/zoobz-io/argus/testing"
)

func TestHookToResponse(t *testing.T) {
	h := argustest.NewHook()
	resp := HookToResponse(h)

	if resp.ID != "h1" || resp.URL != "https://example.com/webhook" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if !resp.Active {
		t.Errorf("Active = %v, want true", resp.Active)
	}
}

func TestHookToCreateResponse(t *testing.T) {
	h := argustest.NewHook()
	resp := HookToCreateResponse(h)

	if resp.ID != "h1" || resp.URL != "https://example.com/webhook" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.Secret != "whsec_test123" {
		t.Errorf("Secret = %q, want %q", resp.Secret, "whsec_test123")
	}
	if !resp.Active {
		t.Errorf("Active = %v, want true", resp.Active)
	}
}

func TestHooksToResponse(t *testing.T) {
	hooks := []*models.Hook{argustest.NewHook(), argustest.NewHook()}
	hooks[1].ID = "h2"

	result := HooksToResponse(hooks)
	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].ID != "h1" || result[1].ID != "h2" {
		t.Errorf("IDs mismatch: %q, %q", result[0].ID, result[1].ID)
	}
}

func TestHooksToResponse_Empty(t *testing.T) {
	result := HooksToResponse(nil)
	if len(result) != 0 {
		t.Errorf("expected empty, got %d", len(result))
	}
}

func TestDeliveryToResponse(t *testing.T) {
	d := argustest.NewDelivery()
	resp := DeliveryToResponse(d)

	if resp.ID != "dl1" || resp.HookID != "h1" || resp.EventID != "evt1" {
		t.Errorf("field mismatch: %+v", resp)
	}
	if resp.StatusCode != 200 || resp.Attempt != 1 {
		t.Errorf("StatusCode/Attempt mismatch: %+v", resp)
	}
}

func TestDeliveryToResponse_WithError(t *testing.T) {
	d := argustest.NewDelivery()
	errMsg := "connection refused"
	d.Error = &errMsg
	d.StatusCode = 0

	resp := DeliveryToResponse(d)
	if resp.Error != "connection refused" {
		t.Errorf("Error = %q, want %q", resp.Error, "connection refused")
	}
}

func TestDeliveriesToListResponse(t *testing.T) {
	d1 := argustest.NewDelivery()
	d2 := argustest.NewDelivery()
	d2.ID = "dl2"

	result := &models.OffsetResult[models.Delivery]{
		Items:  []*models.Delivery{d1, d2},
		Total:  5,
		Offset: 0,
	}

	resp := DeliveriesToListResponse(result)
	if len(resp.Deliveries) != 2 {
		t.Fatalf("expected 2 deliveries, got %d", len(resp.Deliveries))
	}
	if resp.Deliveries[0].ID != "dl1" || resp.Deliveries[1].ID != "dl2" {
		t.Errorf("IDs mismatch: %q, %q", resp.Deliveries[0].ID, resp.Deliveries[1].ID)
	}
	if resp.Total != 5 {
		t.Errorf("Total = %d, want 5", resp.Total)
	}
	if resp.Limit != 2 {
		t.Errorf("Limit = %d, want 2", resp.Limit)
	}
}

func TestDeliveriesToListResponse_Empty(t *testing.T) {
	result := &models.OffsetResult[models.Delivery]{
		Items: []*models.Delivery{},
		Total: 0,
	}
	resp := DeliveriesToListResponse(result)
	if len(resp.Deliveries) != 0 {
		t.Errorf("expected empty, got %d", len(resp.Deliveries))
	}
}
