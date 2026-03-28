package config

import "testing"

func TestLLM_Validate_Valid(t *testing.T) {
	c := LLM{APIKey: "key", Model: "gpt-4"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestLLM_Validate_MissingAPIKey(t *testing.T) {
	c := LLM{Model: "gpt-4"}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing api key")
	}
}

func TestLLM_Validate_MissingModel(t *testing.T) {
	c := LLM{APIKey: "key"}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing model")
	}
}
