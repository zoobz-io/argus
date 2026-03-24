package config

import "errors"

// Embedding holds embedding provider configuration for vector generation.
type Embedding struct {
	Provider   string `env:"APP_EMBEDDING_PROVIDER" default:"openai"`
	APIKey     string `env:"APP_EMBEDDING_API_KEY" secret:"app/embedding-api-key"`
	Model      string `env:"APP_EMBEDDING_MODEL" default:"text-embedding-3-small"`
	BaseURL    string `env:"APP_EMBEDDING_BASE_URL" default:"https://api.openai.com/v1"`
	Dimensions int    `env:"APP_EMBEDDING_DIMENSIONS" default:"1536"`
}

// Validate checks that the configuration is valid.
func (c Embedding) Validate() error {
	if c.APIKey == "" {
		return errors.New("embedding api key is required")
	}
	if c.Model == "" {
		return errors.New("embedding model is required")
	}
	if c.Dimensions <= 0 {
		return errors.New("embedding dimensions must be positive")
	}
	return nil
}
