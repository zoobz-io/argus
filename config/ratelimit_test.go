package config

import "testing"

func TestRateLimit_Validate_Valid(t *testing.T) {
	c := RateLimit{RequestsPerMinute: 120}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestRateLimit_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		rpm  int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := RateLimit{RequestsPerMinute: tt.rpm}
			if err := c.Validate(); err == nil {
				t.Error("expected error for invalid requests per minute")
			}
		})
	}
}
