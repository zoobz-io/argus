package config

import "errors"

// OCR holds OCR gRPC service configuration.
type OCR struct {
	Addr string `env:"APP_OCR_ADDR" default:"localhost:50051"`
}

// Validate checks that the configuration is valid.
func (c OCR) Validate() error {
	if c.Addr == "" {
		return errors.New("ocr service address is required")
	}
	return nil
}
