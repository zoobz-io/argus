package models

import (
	"testing"
	"time"
)

func TestDelivery_GetID(t *testing.T) {
	d := Delivery{ID: "d-1"}
	if got := d.GetID(); got != "d-1" {
		t.Errorf("GetID() = %q, want %q", got, "d-1")
	}
}

func TestDelivery_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	d := Delivery{CreatedAt: ts}
	if got := d.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestDelivery_Clone(t *testing.T) {
	errMsg := "timeout"
	d := Delivery{
		ID:         "d-1",
		HookID:     "h-1",
		EventID:    "evt-1",
		TenantID:   "t-1",
		StatusCode: 200,
		Attempt:    1,
		Error:      &errMsg,
	}
	clone := d.Clone()

	if clone.ID != d.ID || clone.HookID != d.HookID || clone.StatusCode != d.StatusCode || clone.Attempt != d.Attempt {
		t.Error("Clone did not copy all fields")
	}

	if clone.Error == nil || *clone.Error != "timeout" {
		t.Error("Clone did not copy Error pointer")
	}

	*clone.Error = "mutated"
	if *d.Error != "timeout" {
		t.Error("mutating clone Error leaked to original")
	}
}

func TestDelivery_Clone_NilError(t *testing.T) {
	d := Delivery{ID: "d-2"}
	clone := d.Clone()
	if clone.Error != nil {
		t.Error("expected nil Error on clone")
	}
}
