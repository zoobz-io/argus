package config

import "testing"

func TestNotifier_Validate_Valid(t *testing.T) {
	c := Notifier{ConsumerGroup: "argus-notifier"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestNotifier_Validate_MissingConsumerGroup(t *testing.T) {
	c := Notifier{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing consumer group")
	}
}
