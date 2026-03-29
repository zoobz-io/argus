package config

import "errors"

// Connector holds connector service configuration.
type Connector struct {
	PollInterval string `env:"APP_CONNECTOR_POLL_INTERVAL" default:"60s"`
}

// Validate checks that the configuration is valid.
func (c Connector) Validate() error {
	if c.PollInterval == "" {
		return errors.New("poll interval is required")
	}
	return nil
}
