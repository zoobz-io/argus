package provider

import (
	"context"
	"io"
	"net/http"
	"sort"
	"testing"
)

// stubProvider is a minimal Provider implementation for registry tests.
type stubProvider struct {
	providerType string
}

func (s *stubProvider) Type() string { return s.providerType }
func (s *stubProvider) AuthURL(_ context.Context, _, _ string) (string, error) {
	return "", nil
}
func (s *stubProvider) Exchange(_ context.Context, _, _ string) (*Credentials, error) {
	return nil, nil
}
func (s *stubProvider) List(_ context.Context, _ *Credentials, _ string) ([]Entry, *Credentials, error) {
	return nil, nil, nil
}
func (s *stubProvider) Changes(_ context.Context, _ *Credentials, _, _ string) ([]Change, string, *Credentials, error) {
	return nil, "", nil, nil
}
func (s *stubProvider) Fetch(_ context.Context, _ *Credentials, _ string) (io.ReadCloser, *EntryMeta, *Credentials, error) {
	return nil, nil, nil, nil
}

// Compile-time check that stubProvider satisfies Provider.
var _ Provider = (*stubProvider)(nil)

// stubWebhookProvider satisfies both Provider and WebhookProvider.
type stubWebhookProvider struct {
	stubProvider
}

func (s *stubWebhookProvider) RegisterWebhook(_ context.Context, _ *Credentials, _, _ string) (*WebhookRegistration, *Credentials, error) {
	return nil, nil, nil
}
func (s *stubWebhookProvider) RenewWebhook(_ context.Context, _ *Credentials, _ *WebhookRegistration) (*WebhookRegistration, *Credentials, error) {
	return nil, nil, nil
}
func (s *stubWebhookProvider) VerifyWebhook(_ *http.Request) (string, error) {
	return "", nil
}

var _ WebhookProvider = (*stubWebhookProvider)(nil)

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{providerType: "google_drive"})

	p, err := r.Get("google_drive")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.Type() != "google_drive" {
		t.Errorf("type: got %q, want %q", p.Type(), "google_drive")
	}
}

func TestRegistry_Get_NotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Get("nonexistent")
	if err == nil {
		t.Fatal("expected error for unregistered type")
	}
}

func TestRegistry_Register_Duplicate_Panics(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{providerType: "google_drive"})

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic on duplicate registration")
		}
	}()
	r.Register(&stubProvider{providerType: "google_drive"})
}

func TestRegistry_Types(t *testing.T) {
	r := NewRegistry()
	r.Register(&stubProvider{providerType: "google_drive"})
	r.Register(&stubProvider{providerType: "onedrive"})
	r.Register(&stubProvider{providerType: "dropbox"})

	types := r.Types()
	sort.Strings(types)
	expected := []string{"dropbox", "google_drive", "onedrive"}
	if len(types) != len(expected) {
		t.Fatalf("types: got %v, want %v", types, expected)
	}
	for i, typ := range types {
		if typ != expected[i] {
			t.Errorf("types[%d]: got %q, want %q", i, typ, expected[i])
		}
	}
}

func TestWebhookProvider_TypeAssertion(t *testing.T) {
	var p Provider = &stubWebhookProvider{stubProvider{providerType: "google_drive"}}

	wp, ok := p.(WebhookProvider)
	if !ok {
		t.Fatal("expected stubWebhookProvider to satisfy WebhookProvider")
	}
	if wp == nil {
		t.Fatal("webhook provider should not be nil")
	}
}

func TestWebhookProvider_TypeAssertion_NotSupported(t *testing.T) {
	var p Provider = &stubProvider{providerType: "s3"}

	_, ok := p.(WebhookProvider)
	if ok {
		t.Fatal("stubProvider should not satisfy WebhookProvider")
	}
}
