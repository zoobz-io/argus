package config

import "testing"

func validEmbedding() Embedding {
	return Embedding{APIKey: "key", Model: "text-embedding-3-small", Dimensions: 1536}
}

func TestEmbedding_Validate_Valid(t *testing.T) {
	if err := validEmbedding().Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestEmbedding_Validate_MissingAPIKey(t *testing.T) {
	c := validEmbedding()
	c.APIKey = ""
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing api key")
	}
}

func TestEmbedding_Validate_MissingModel(t *testing.T) {
	c := validEmbedding()
	c.Model = ""
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing model")
	}
}

func TestEmbedding_Validate_InvalidDimensions(t *testing.T) {
	for _, dim := range []int{0, -1} {
		c := validEmbedding()
		c.Dimensions = dim
		if err := c.Validate(); err == nil {
			t.Errorf("expected error for dimensions %d", dim)
		}
	}
}
