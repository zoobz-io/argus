package config

import (
	"errors"
	"time"
)

// Connector holds connector service configuration.
type Connector struct {
	PollInterval time.Duration `env:"APP_CONNECTOR_POLL_INTERVAL" default:"60s"`
	DrainTimeout time.Duration `env:"APP_CONNECTOR_DRAIN_TIMEOUT" default:"30s"`
}

// Validate checks that the configuration is valid.
func (c Connector) Validate() error {
	if c.PollInterval <= 0 {
		return errors.New("poll interval must be positive")
	}
	if c.DrainTimeout <= 0 {
		return errors.New("drain timeout must be positive")
	}
	return nil
}
