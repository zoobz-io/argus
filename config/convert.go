package config

import (
	"errors"
	"time"
)

// Convert holds document conversion gRPC service configuration.
type Convert struct {
	Addr    string        `env:"APP_CONVERT_ADDR" default:"localhost:50052"`
	Timeout time.Duration `env:"APP_CONVERT_TIMEOUT" default:"120s"`
}

// Validate checks that the configuration is valid.
func (c Convert) Validate() error {
	if c.Addr == "" {
		return errors.New("convert service address is required")
	}
	if c.Timeout <= 0 {
		return errors.New("convert timeout must be positive")
	}
	return nil
}
