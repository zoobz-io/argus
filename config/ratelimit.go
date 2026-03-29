package config

import "errors"

// RateLimit holds API rate limiting configuration.
type RateLimit struct {
	RequestsPerMinute int  `env:"APP_RATE_LIMIT_RPM" default:"120"`
	TrustProxy        bool `env:"APP_RATE_LIMIT_TRUST_PROXY" default:"false"`
}

// Validate checks that the configuration is valid.
func (c RateLimit) Validate() error {
	if c.RequestsPerMinute <= 0 {
		return errors.New("requests per minute must be positive")
	}
	return nil
}
