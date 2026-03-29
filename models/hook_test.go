//go:build testing

package models

import (
	"context"
	"encoding/hex"
	"testing"
	"time"

	"github.com/zoobz-io/cereal"
	"github.com/zoobz-io/sum"
)

func TestHook_GetID(t *testing.T) {
	h := Hook{ID: "h-1"}
	if got := h.GetID(); got != "h-1" {
		t.Errorf("GetID() = %q, want %q", got, "h-1")
	}
}

func TestHook_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	h := Hook{CreatedAt: ts}
	if got := h.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestHook_Clone(t *testing.T) {
	h := Hook{ID: "h-1", URL: "https://example.com/hook", Secret: "secret", Active: true, TenantID: "t-1", UserID: "u-1"}
	clone := h.Clone()

	if clone.ID != h.ID || clone.URL != h.URL || clone.Secret != h.Secret || clone.Active != h.Active {
		t.Error("Clone did not copy all fields")
	}

	clone.URL = "mutated"
	if h.URL != "https://example.com/hook" {
		t.Error("mutating clone leaked to original")
	}
}

func setupHookBoundary(t *testing.T) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)

	keyHex := "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	key, err := hex.DecodeString(keyHex)
	if err != nil {
		t.Fatalf("decoding key: %v", err)
	}
	enc, err := cereal.AES(key)
	if err != nil {
		t.Fatalf("creating encryptor: %v", err)
	}

	svc := sum.New()
	svc.WithEncryptor(cereal.EncryptAES, enc)
	k := sum.Start()
	sum.NewBoundary[Hook](k)
	sum.Freeze(k)
	return context.Background()
}

func setupHookBoundaryBroken(t *testing.T) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)

	svc := sum.New()
	svc.WithEncryptor(cereal.EncryptAES, &hookFailEncryptor{})
	k := sum.Start()
	sum.NewBoundary[Hook](k)
	sum.Freeze(k)
	return context.Background()
}

type hookFailEncryptor struct{}

func (f *hookFailEncryptor) Encrypt(_ []byte) ([]byte, error) {
	return nil, cereal.ErrMissingEncryptor
}

func (f *hookFailEncryptor) Decrypt(_ []byte) ([]byte, error) {
	return nil, cereal.ErrMissingEncryptor
}

func TestHook_BeforeSave_Error(t *testing.T) {
	ctx := setupHookBoundaryBroken(t)
	h := &Hook{ID: "h-1", Secret: "secret"}
	err := h.BeforeSave(ctx)
	if err == nil {
		t.Fatal("expected error from broken encryptor")
	}
}

func TestHook_AfterLoad_Error(t *testing.T) {
	ctx := setupHookBoundaryBroken(t)
	h := &Hook{ID: "h-1", Secret: "not-valid-ciphertext"}
	err := h.AfterLoad(ctx)
	if err == nil {
		t.Fatal("expected error from broken encryptor")
	}
}

func TestHook_BeforeSave_AfterLoad_RoundTrip(t *testing.T) {
	ctx := setupHookBoundary(t)
	h := &Hook{
		ID:     "h-1",
		URL:    "https://example.com/hook",
		Secret: "whsec_test123",
	}

	err := h.BeforeSave(ctx)
	if err != nil {
		t.Fatalf("BeforeSave error: %v", err)
	}
	if h.Secret == "whsec_test123" {
		t.Error("Secret should be encrypted after BeforeSave")
	}
	if h.ID != "h-1" {
		t.Errorf("ID changed: got %q", h.ID)
	}

	err = h.AfterLoad(ctx)
	if err != nil {
		t.Fatalf("AfterLoad error: %v", err)
	}
	if h.Secret != "whsec_test123" {
		t.Errorf("Secret after round-trip: got %q, want %q", h.Secret, "whsec_test123")
	}
}
