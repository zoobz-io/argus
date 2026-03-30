package config

import (
	"testing"
	"time"
)

func TestNotifier_Validate_Valid(t *testing.T) {
	c := Notifier{ConsumerGroup: "argus-notifier", DrainTimeout: 30 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestNotifier_Validate_MissingConsumerGroup(t *testing.T) {
	c := Notifier{DrainTimeout: 30 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing consumer group")
	}
}

func TestNotifier_Validate_InvalidDrainTimeout(t *testing.T) {
	c := Notifier{ConsumerGroup: "argus-notifier", DrainTimeout: 0}
	if err := c.Validate(); err == nil {
		t.Error("expected error for zero drain timeout")
	}
}
