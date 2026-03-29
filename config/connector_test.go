package config

import "testing"

func TestConnector_Validate_Valid(t *testing.T) {
	c := Connector{PollInterval: "60s"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestConnector_Validate_MissingInterval(t *testing.T) {
	c := Connector{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing poll interval")
	}
}
