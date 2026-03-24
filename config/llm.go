package config

import "errors"

// LLM holds LLM provider configuration for content summarization.
type LLM struct {
	Provider string `env:"APP_LLM_PROVIDER" default:"openai"`
	APIKey   string `env:"APP_LLM_API_KEY" secret:"app/llm-api-key"`
	Model    string `env:"APP_LLM_MODEL" default:"gpt-4"`
	BaseURL  string `env:"APP_LLM_BASE_URL" default:"https://api.openai.com/v1"`
}

// Validate checks that the configuration is valid.
func (c LLM) Validate() error {
	if c.APIKey == "" {
		return errors.New("llm api key is required")
	}
	if c.Model == "" {
		return errors.New("llm model is required")
	}
	return nil
}
