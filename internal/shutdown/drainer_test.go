package shutdown

import (
	"sync"
	"testing"
	"time"
)

func TestDrainer_Track_And_Drain(t *testing.T) {
	var d Drainer

	done1 := d.Track("job-1")
	done2 := d.Track("job-2")

	// Complete both before draining.
	done1()
	done2()

	interrupted := d.Drain(time.Second)
	if len(interrupted) != 0 {
		t.Errorf("expected clean drain, got interrupted: %v", interrupted)
	}
}

func TestDrainer_Drain_Timeout(t *testing.T) {
	var d Drainer

	// Track an operation that never completes.
	_ = d.Track("stuck-job")

	interrupted := d.Drain(50 * time.Millisecond)
	if len(interrupted) != 1 || interrupted[0] != "stuck-job" {
		t.Errorf("expected [stuck-job], got %v", interrupted)
	}
}

func TestDrainer_Drain_Partial(t *testing.T) {
	var d Drainer

	done1 := d.Track("fast-job")
	_ = d.Track("slow-job")

	// Complete only one.
	done1()

	interrupted := d.Drain(50 * time.Millisecond)
	if len(interrupted) != 1 || interrupted[0] != "slow-job" {
		t.Errorf("expected [slow-job], got %v", interrupted)
	}
}

func TestDrainer_Concurrent_Track(t *testing.T) {
	var d Drainer
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			done := d.Track("concurrent")
			time.Sleep(time.Millisecond)
			done()
		}()
	}
	wg.Wait()

	interrupted := d.Drain(time.Second)
	if len(interrupted) != 0 {
		t.Errorf("expected clean drain, got %v", interrupted)
	}
}

func TestDrainer_ZeroValue(t *testing.T) {
	// Zero-value Drainer should work without initialization.
	var d Drainer
	interrupted := d.Drain(50 * time.Millisecond)
	if len(interrupted) != 0 {
		t.Errorf("expected empty, got %v", interrupted)
	}
}
