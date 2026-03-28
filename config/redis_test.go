package config

import "testing"

func TestRedis_Validate_Valid(t *testing.T) {
	c := Redis{Addr: "localhost:6379"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestRedis_Validate_MissingAddr(t *testing.T) {
	c := Redis{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing addr")
	}
}
