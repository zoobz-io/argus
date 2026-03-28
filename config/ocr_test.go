package config

import (
	"testing"
	"time"
)

func TestOCR_Validate_Valid(t *testing.T) {
	c := OCR{Addr: "localhost:50051", Timeout: 120 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestOCR_Validate_MissingAddr(t *testing.T) {
	c := OCR{Timeout: 120 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing addr")
	}
}

func TestOCR_Validate_ZeroTimeout(t *testing.T) {
	c := OCR{Addr: "localhost:50051"}
	if err := c.Validate(); err == nil {
		t.Error("expected error for zero timeout")
	}
}
