package config

import "errors"

// Notifier holds notification sidecar configuration.
type Notifier struct {
	ConsumerGroup string `env:"APP_NOTIFIER_CONSUMER_GROUP" default:"argus-notifier"`
}

// Validate checks that the configuration is valid.
func (c Notifier) Validate() error {
	if c.ConsumerGroup == "" {
		return errors.New("consumer group is required")
	}
	return nil
}
