package config

import (
	"errors"
	"time"
)

// Connector holds connector service configuration.
type Connector struct {
	PollInterval time.Duration `env:"APP_CONNECTOR_POLL_INTERVAL" default:"60s"`
}

// Validate checks that the configuration is valid.
func (c Connector) Validate() error {
	if c.PollInterval <= 0 {
		return errors.New("poll interval must be positive")
	}
	return nil
}
