package config

import "errors"

// Convert holds document conversion gRPC service configuration.
type Convert struct {
	Addr string `env:"APP_CONVERT_ADDR" default:"localhost:50052"`
}

// Validate checks that the configuration is valid.
func (c Convert) Validate() error {
	if c.Addr == "" {
		return errors.New("convert service address is required")
	}
	return nil
}
