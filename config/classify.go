package config

import (
	"errors"
	"time"
)

// Classify holds prompt injection classifier gRPC service configuration.
type Classify struct {
	Addr    string        `env:"APP_CLASSIFY_ADDR" default:"localhost:50053"`
	Timeout time.Duration `env:"APP_CLASSIFY_TIMEOUT" default:"30s"`
}

// Validate checks that the configuration is valid.
func (c Classify) Validate() error {
	if c.Addr == "" {
		return errors.New("classify service address is required")
	}
	if c.Timeout <= 0 {
		return errors.New("classify timeout must be positive")
	}
	return nil
}
