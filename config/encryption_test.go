package config

import (
	"strings"
	"testing"
)

func TestEncryption_Validate_Valid(t *testing.T) {
	c := Encryption{Key: strings.Repeat("ab", 32)} // 64 hex chars
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestEncryption_Validate_Empty(t *testing.T) {
	c := Encryption{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for empty key")
	}
}

func TestEncryption_Validate_WrongLength(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"too short", strings.Repeat("a", 63)},
		{"too long", strings.Repeat("a", 65)},
		{"half length", strings.Repeat("a", 32)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Encryption{Key: tt.key}
			if err := c.Validate(); err == nil {
				t.Errorf("expected error for key length %d", len(tt.key))
			}
		})
	}
}
