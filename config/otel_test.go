package config

import "testing"

func TestOTEL_Validate_Valid(t *testing.T) {
	c := OTEL{Endpoint: "localhost:4318"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestOTEL_Validate_MissingEndpoint(t *testing.T) {
	c := OTEL{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing endpoint")
	}
}
