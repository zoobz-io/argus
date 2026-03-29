//go:build testing

package notify

import (
	"testing"
)

func TestNewWebhookPipeline(t *testing.T) {
	p := NewWebhookPipeline()
	if p == nil {
		t.Fatal("expected non-nil pipeline")
	}
	if p.sequence == nil {
		t.Fatal("expected non-nil sequence")
	}
}
