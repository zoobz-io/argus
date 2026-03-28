package boot

import "testing"

func TestHostname_ReturnsNonEmpty(t *testing.T) {
	h := Hostname()
	if h == "" {
		t.Error("hostname should not be empty")
	}
}

func TestHostname_Deterministic(t *testing.T) {
	// On a real system, os.Hostname succeeds and returns the same value.
	h1 := Hostname()
	h2 := Hostname()
	if h1 != h2 {
		t.Errorf("hostname should be deterministic, got %q and %q", h1, h2)
	}
}
