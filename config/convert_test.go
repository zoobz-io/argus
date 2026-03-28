package config

import (
	"testing"
	"time"
)

func TestConvert_Validate_Valid(t *testing.T) {
	c := Convert{Addr: "localhost:50052", Timeout: 120 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestConvert_Validate_MissingAddr(t *testing.T) {
	c := Convert{Timeout: 120 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing addr")
	}
}

func TestConvert_Validate_ZeroTimeout(t *testing.T) {
	c := Convert{Addr: "localhost:50052"}
	if err := c.Validate(); err == nil {
		t.Error("expected error for zero timeout")
	}
}
