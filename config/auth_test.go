package config

import "testing"

func TestAuth_Validate_Valid(t *testing.T) {
	c := Auth{Issuer: "http://localhost:8085"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestAuth_Validate_MissingIssuer(t *testing.T) {
	c := Auth{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing issuer")
	}
}
