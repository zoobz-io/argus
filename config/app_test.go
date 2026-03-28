package config

import "testing"

func TestApp_Validate_Valid(t *testing.T) {
	c := App{Port: 8080}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestApp_Validate_Invalid(t *testing.T) {
	tests := []struct {
		name string
		port int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := App{Port: tt.port}
			if err := c.Validate(); err == nil {
				t.Error("expected error")
			}
		})
	}
}
