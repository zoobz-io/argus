package wire

import "testing"

func TestURLMasker_ValidURL(t *testing.T) {
	m := URLMasker()
	result, err := m.Mask("https://example.com/webhook/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://e***/***" {
		t.Errorf("masked = %q, want https://e***/***", result)
	}
}

func TestURLMasker_InvalidURL(t *testing.T) {
	m := URLMasker()
	result, err := m.Mask("not-a-url")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "***" {
		t.Errorf("masked = %q, want ***", result)
	}
}

func TestURLMasker_WithPort(t *testing.T) {
	m := URLMasker()
	result, err := m.Mask("https://example.com:8443/path")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "https://e***:8443/***" {
		t.Errorf("masked = %q, want https://e***:8443/***", result)
	}
}

func TestSecretMasker(t *testing.T) {
	m := SecretMasker()
	result, err := m.Mask("whsec_abc123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "************" {
		t.Errorf("masked = %q, want 12 asterisks", result)
	}
}

func TestSecretMasker_Empty(t *testing.T) {
	m := SecretMasker()
	result, err := m.Mask("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "***" {
		t.Errorf("masked = %q, want ***", result)
	}
}
