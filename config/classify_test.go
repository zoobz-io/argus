package config

import "testing"

func TestClassify_Validate_Valid(t *testing.T) {
	c := Classify{Addr: "localhost:50053"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestClassify_Validate_MissingAddr(t *testing.T) {
	c := Classify{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing addr")
	}
}
