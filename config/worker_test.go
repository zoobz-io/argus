package config

import "testing"

func TestWorker_Validate_Valid(t *testing.T) {
	c := Worker{ConsumerGroup: "argus-worker"}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestWorker_Validate_MissingConsumerGroup(t *testing.T) {
	c := Worker{}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing consumer group")
	}
}
