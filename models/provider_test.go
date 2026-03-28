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

func TestProvider_GetID(t *testing.T) {
	p := Provider{ID: "p-1"}
	if got := p.GetID(); got != "p-1" {
		t.Errorf("GetID() = %q, want %q", got, "p-1")
	}
}

func TestProvider_GetCreatedAt(t *testing.T) {
	ts := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	p := Provider{CreatedAt: ts}
	if got := p.GetCreatedAt(); !got.Equal(ts) {
		t.Errorf("GetCreatedAt() = %v, want %v", got, ts)
	}
}

func TestProvider_Clone(t *testing.T) {
	p := Provider{ID: "p-1", Name: "My Drive", Type: ProviderGoogleDrive, Active: true, Credentials: "secret"}
	clone := p.Clone()

	if clone.ID != p.ID || clone.Name != p.Name || clone.Type != p.Type || clone.Credentials != p.Credentials {
		t.Error("Clone did not copy all fields")
	}

	clone.Name = "mutated"
	if p.Name != "My Drive" {
		t.Error("mutating clone leaked to original")
	}
}

func setupProviderBoundary(t *testing.T) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)

	// 32-byte AES key (hex-encoded 64 chars, decoded to 32 bytes).
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
	sum.NewBoundary[Provider](k)
	sum.Freeze(k)
	return context.Background()
}

func setupProviderBoundaryBroken(t *testing.T) context.Context {
	t.Helper()
	sum.Reset()
	t.Cleanup(sum.Reset)

	// Register a broken encryptor that always fails.
	svc := sum.New()
	svc.WithEncryptor(cereal.EncryptAES, &failEncryptor{})
	k := sum.Start()
	sum.NewBoundary[Provider](k)
	sum.Freeze(k)
	return context.Background()
}

type failEncryptor struct{}

func (f *failEncryptor) Encrypt(_ []byte) ([]byte, error) {
	return nil, cereal.ErrMissingEncryptor
}

func (f *failEncryptor) Decrypt(_ []byte) ([]byte, error) {
	return nil, cereal.ErrMissingEncryptor
}

func TestProvider_BeforeSave_Error(t *testing.T) {
	ctx := setupProviderBoundaryBroken(t)
	p := &Provider{ID: "p-1", Credentials: "secret"}
	err := p.BeforeSave(ctx)
	if err == nil {
		t.Fatal("expected error from broken encryptor")
	}
}

func TestProvider_AfterLoad_Error(t *testing.T) {
	ctx := setupProviderBoundaryBroken(t)
	p := &Provider{ID: "p-1", Credentials: "not-valid-ciphertext"}
	err := p.AfterLoad(ctx)
	if err == nil {
		t.Fatal("expected error from broken encryptor")
	}
}

func TestProvider_BeforeSave_AfterLoad_RoundTrip(t *testing.T) {
	ctx := setupProviderBoundary(t)
	p := &Provider{
		ID:          "p-1",
		Name:        "Test",
		Credentials: "secret-token",
	}

	// BeforeSave encrypts credentials.
	err := p.BeforeSave(ctx)
	if err != nil {
		t.Fatalf("BeforeSave error: %v", err)
	}
	if p.Credentials == "secret-token" {
		t.Error("Credentials should be encrypted after BeforeSave")
	}
	if p.ID != "p-1" {
		t.Errorf("ID changed: got %q", p.ID)
	}

	// AfterLoad decrypts credentials.
	err = p.AfterLoad(ctx)
	if err != nil {
		t.Fatalf("AfterLoad error: %v", err)
	}
	if p.Credentials != "secret-token" {
		t.Errorf("Credentials after round-trip: got %q, want %q", p.Credentials, "secret-token")
	}
}
