package config

import (
	"testing"
	"time"
)

func TestConnector_Validate_Valid(t *testing.T) {
	c := Connector{PollInterval: 60 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestConnector_Validate_ZeroInterval(t *testing.T) {
	c := Connector{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for zero poll interval")
	}
}

func TestConnector_Validate_NegativeInterval(t *testing.T) {
	c := Connector{PollInterval: -1 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for negative poll interval")
	}
}
