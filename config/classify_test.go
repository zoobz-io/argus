package config

import (
	"testing"
	"time"
)

func TestClassify_Validate_Valid(t *testing.T) {
	c := Classify{Addr: "localhost:50053", Timeout: 30 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestClassify_Validate_MissingAddr(t *testing.T) {
	c := Classify{Timeout: 30 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing addr")
	}
}

func TestClassify_Validate_ZeroTimeout(t *testing.T) {
	c := Classify{Addr: "localhost:50053"}
	if err := c.Validate(); err == nil {
		t.Error("expected error for zero timeout")
	}
}
