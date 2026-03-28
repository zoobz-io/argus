package config

import "errors"

// Auth holds OIDC authentication configuration.
type Auth struct {
	Issuer   string `env:"APP_AUTH_ISSUER" default:"http://localhost:8085"`
	Audience string `env:"APP_AUTH_AUDIENCE"`
}

// Validate checks that the configuration is valid.
func (c Auth) Validate() error {
	if c.Issuer == "" {
		return errors.New("auth issuer URL is required")
	}
	return nil
}
