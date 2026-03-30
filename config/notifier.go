package config

import (
	"errors"
	"time"
)

// Notifier holds notification sidecar configuration.
type Notifier struct {
	ConsumerGroup string        `env:"APP_NOTIFIER_CONSUMER_GROUP" default:"argus-notifier"`
	DrainTimeout  time.Duration `env:"APP_NOTIFIER_DRAIN_TIMEOUT" default:"30s"`
	FanOutWorkers int           `env:"APP_NOTIFIER_FANOUT_WORKERS" default:"4"`
}

// Validate checks that the configuration is valid.
func (c Notifier) Validate() error {
	if c.ConsumerGroup == "" {
		return errors.New("consumer group is required")
	}
	if c.DrainTimeout <= 0 {
		return errors.New("drain timeout must be positive")
	}
	if c.FanOutWorkers <= 0 {
		return errors.New("fan-out workers must be positive")
	}
	return nil
}
