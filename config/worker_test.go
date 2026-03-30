package config

import (
	"testing"
	"time"
)

func TestWorker_Validate_Valid(t *testing.T) {
	c := Worker{ConsumerGroup: "argus-worker", WorkerCount: 4, DrainTimeout: 30 * time.Second}
	if err := c.Validate(); err != nil {
		t.Errorf("expected valid, got: %v", err)
	}
}

func TestWorker_Validate_MissingConsumerGroup(t *testing.T) {
	c := Worker{WorkerCount: 4, DrainTimeout: 30 * time.Second}
	if err := c.Validate(); err == nil {
		t.Error("expected error for missing consumer group")
	}
}

func TestWorker_Validate_InvalidWorkerCount(t *testing.T) {
	tests := []struct {
		name  string
		count int
	}{
		{"zero", 0},
		{"negative", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Worker{ConsumerGroup: "argus-worker", WorkerCount: tt.count, DrainTimeout: 30 * time.Second}
			if err := c.Validate(); err == nil {
				t.Error("expected error for invalid worker count")
			}
		})
	}
}

func TestWorker_Validate_InvalidDrainTimeout(t *testing.T) {
	c := Worker{ConsumerGroup: "argus-worker", WorkerCount: 4, DrainTimeout: 0}
	if err := c.Validate(); err == nil {
		t.Error("expected error for zero drain timeout")
	}
}
