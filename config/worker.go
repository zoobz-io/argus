package config

import "errors"

// Worker holds worker sidecar configuration.
type Worker struct {
	ConsumerGroup string `env:"APP_WORKER_CONSUMER_GROUP" default:"argus-worker"`
}

// Validate checks that the configuration is valid.
func (c Worker) Validate() error {
	if c.ConsumerGroup == "" {
		return errors.New("consumer group is required")
	}
	return nil
}
