package config

import "errors"

// Classify holds prompt injection classifier gRPC service configuration.
type Classify struct {
	Addr string `env:"APP_CLASSIFY_ADDR" default:"localhost:50053"`
}

// Validate checks that the configuration is valid.
func (c Classify) Validate() error {
	if c.Addr == "" {
		return errors.New("classify service address is required")
	}
	return nil
}
