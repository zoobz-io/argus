// Package shutdown provides graceful drain utilities for sidecar services.
//
// On SIGTERM, services should stop accepting new work (cancel the subscriber
// context) then call Drainer.Drain to wait for in-flight operations to finish
// within a bounded timeout.
package shutdown

import (
	"log"
	"sync"
	"time"
)

// Drainer tracks in-flight operations and waits for them to complete on
// shutdown with a bounded timeout. Zero value is ready to use.
type Drainer struct {
	active map[string]time.Time
	wg     sync.WaitGroup
	mu     sync.Mutex
}

// Track registers an in-flight operation by ID. The returned function must
// be called when the operation completes (typically via defer).
func (d *Drainer) Track(id string) func() {
	d.mu.Lock()
	if d.active == nil {
		d.active = make(map[string]time.Time)
	}
	d.active[id] = time.Now()
	d.mu.Unlock()
	d.wg.Add(1)

	return func() {
		d.mu.Lock()
		delete(d.active, id)
		d.mu.Unlock()
		d.wg.Done()
	}
}

// Drain waits for all tracked operations to complete or the timeout to expire.
// Returns the IDs of operations that were still in-flight when the timeout
// expired. An empty slice means all work drained cleanly.
func (d *Drainer) Drain(timeout time.Duration) []string {
	done := make(chan struct{})
	go func() {
		d.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("shutdown: all in-flight operations drained")
		return nil
	case <-time.After(timeout):
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	interrupted := make([]string, 0, len(d.active))
	for id := range d.active {
		interrupted = append(interrupted, id)
	}
	log.Printf("shutdown: drain timeout exceeded, %d operations interrupted: %v", len(interrupted), interrupted)
	return interrupted
}
