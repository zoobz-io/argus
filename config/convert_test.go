package config

import "testing"

func TestConvert_Validate_Valid(t *testing.T) {
	c := Convert{Addr: "localhost:50052"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestConvert_Validate_MissingAddr(t *testing.T) {
	c := Convert{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing addr")
	}
}
