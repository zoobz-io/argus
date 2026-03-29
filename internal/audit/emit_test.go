package audit

import (
	"context"
	"testing"
)

func TestEmit_DoesNotPanic(t *testing.T) {
	ctx := context.Background()

	// Verify Emit does not panic with valid metadata.
	Emit(ctx, "provider.created", "provider", "p-1", "t-1", "u-1", map[string]any{
		"provider_type": "google_drive",
	})
}

func TestEmit_NilMetadata(t *testing.T) {
	ctx := context.Background()

	// Verify Emit does not panic with nil metadata.
	Emit(ctx, "tenant.deleted", "tenant", "t-1", "t-1", "admin", nil)
}
