package config

import (
	"errors"
	"time"
)

// Worker holds worker sidecar configuration.
type Worker struct {
	ConsumerGroup string        `env:"APP_WORKER_CONSUMER_GROUP" default:"argus-worker"`
	WorkerCount   int           `env:"APP_WORKER_COUNT" default:"4"`
	DrainTimeout  time.Duration `env:"APP_WORKER_DRAIN_TIMEOUT" default:"30s"`
}

// Validate checks that the configuration is valid.
func (c Worker) Validate() error {
	if c.ConsumerGroup == "" {
		return errors.New("consumer group is required")
	}
	if c.WorkerCount <= 0 {
		return errors.New("worker count must be positive")
	}
	if c.DrainTimeout <= 0 {
		return errors.New("drain timeout must be positive")
	}
	return nil
}
