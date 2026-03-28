package config

import (
	"errors"
	"time"
)

// OCR holds OCR gRPC service configuration.
type OCR struct {
	Addr    string        `env:"APP_OCR_ADDR" default:"localhost:50051"`
	Timeout time.Duration `env:"APP_OCR_TIMEOUT" default:"120s"`
}

// Validate checks that the configuration is valid.
func (c OCR) Validate() error {
	if c.Addr == "" {
		return errors.New("ocr service address is required")
	}
	if c.Timeout <= 0 {
		return errors.New("ocr timeout must be positive")
	}
	return nil
}
