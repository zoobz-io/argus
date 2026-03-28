package provider

import (
	"testing"
	"time"
)

func TestCredentials_Expired_NotExpired(t *testing.T) {
	c := &Credentials{
		Expiry: time.Now().Add(time.Hour),
	}
	if c.Expired() {
		t.Error("expected not expired")
	}
}

func TestCredentials_Expired_Expired(t *testing.T) {
	c := &Credentials{
		Expiry: time.Now().Add(-time.Minute),
	}
	if !c.Expired() {
		t.Error("expected expired")
	}
}

func TestCredentials_Expired_WithinBuffer(t *testing.T) {
	// Within the 30s buffer — should report expired.
	c := &Credentials{
		Expiry: time.Now().Add(15 * time.Second),
	}
	if !c.Expired() {
		t.Error("expected expired within 30s buffer")
	}
}

func TestCredentials_Expired_ZeroExpiry(t *testing.T) {
	c := &Credentials{}
	if c.Expired() {
		t.Error("zero expiry should not be treated as expired")
	}
}
