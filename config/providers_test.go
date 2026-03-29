package config

import "testing"

func TestProviders_Validate(t *testing.T) {
	c := Providers{}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}
