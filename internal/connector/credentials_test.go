package connector

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

type mockProviderStore struct {
	onGetProvider              func(ctx context.Context, id string) (*models.Provider, error)
	onUpdateProviderCredentials func(ctx context.Context, id, credentials string) error
}

func (m *mockProviderStore) GetProvider(ctx context.Context, id string) (*models.Provider, error) {
	if m.onGetProvider != nil {
		return m.onGetProvider(ctx, id)
	}
	return &models.Provider{}, nil
}

func (m *mockProviderStore) UpdateProviderCredentials(ctx context.Context, id, credentials string) error {
	if m.onUpdateProviderCredentials != nil {
		return m.onUpdateProviderCredentials(ctx, id, credentials)
	}
	return nil
}

func credJSON(t *testing.T, creds *provider.Credentials) string {
	t.Helper()
	b, err := json.Marshal(creds)
	if err != nil {
		t.Fatalf("marshal creds: %v", err)
	}
	return string(b)
}

func TestCredentialManager_Get_LoadsFromDB(t *testing.T) {
	creds := &provider.Credentials{AccessToken: "token-1", RefreshToken: "refresh-1"}
	store := &mockProviderStore{
		onGetProvider: func(_ context.Context, id string) (*models.Provider, error) {
			return &models.Provider{ID: id, Credentials: credJSON(t, creds)}, nil
		},
	}

	cm := NewCredentialManager(store)
	got, err := cm.Get(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccessToken != "token-1" {
		t.Errorf("access token: got %q", got.AccessToken)
	}
}

func TestCredentialManager_Get_ReturnsCached(t *testing.T) {
	calls := 0
	store := &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			calls++
			creds := &provider.Credentials{AccessToken: "token-1"}
			return &models.Provider{Credentials: credJSON(t, creds)}, nil
		},
	}

	cm := NewCredentialManager(store)
	_, _ = cm.Get(context.Background(), "p-1")
	_, _ = cm.Get(context.Background(), "p-1")

	if calls != 1 {
		t.Errorf("expected 1 DB call, got %d", calls)
	}
}

func TestCredentialManager_Get_DBError(t *testing.T) {
	store := &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return nil, errors.New("db down")
		},
	}

	cm := NewCredentialManager(store)
	_, err := cm.Get(context.Background(), "p-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCredentialManager_Get_InvalidJSON(t *testing.T) {
	store := &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			return &models.Provider{Credentials: "not json"}, nil
		},
	}

	cm := NewCredentialManager(store)
	_, err := cm.Get(context.Background(), "p-1")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestCredentialManager_Update_PersistsAndCaches(t *testing.T) {
	var storedCreds string
	store := &mockProviderStore{
		onUpdateProviderCredentials: func(_ context.Context, _, creds string) error {
			storedCreds = creds
			return nil
		},
	}

	cm := NewCredentialManager(store)
	creds := &provider.Credentials{AccessToken: "new-token", RefreshToken: "new-refresh"}

	if err := cm.Update(context.Background(), "p-1", creds); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if storedCreds == "" {
		t.Error("expected credentials to be persisted")
	}

	// Should be cached now — no DB call needed.
	cm.store = &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			t.Fatal("should not hit DB after Update cached")
			return nil, nil
		},
	}
	got, err := cm.Get(context.Background(), "p-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AccessToken != "new-token" {
		t.Errorf("cached token: got %q", got.AccessToken)
	}
}

func TestCredentialManager_Update_NilIsNoop(t *testing.T) {
	store := &mockProviderStore{
		onUpdateProviderCredentials: func(_ context.Context, _, _ string) error {
			t.Fatal("should not be called for nil creds")
			return nil
		},
	}

	cm := NewCredentialManager(store)
	if err := cm.Update(context.Background(), "p-1", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestCredentialManager_Update_DBError(t *testing.T) {
	store := &mockProviderStore{
		onUpdateProviderCredentials: func(_ context.Context, _, _ string) error {
			return errors.New("db error")
		},
	}

	cm := NewCredentialManager(store)
	err := cm.Update(context.Background(), "p-1", &provider.Credentials{AccessToken: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCredentialManager_Invalidate(t *testing.T) {
	calls := 0
	store := &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			calls++
			creds := &provider.Credentials{AccessToken: "token-1"}
			return &models.Provider{Credentials: credJSON(t, creds)}, nil
		},
	}

	cm := NewCredentialManager(store)
	_, _ = cm.Get(context.Background(), "p-1")
	cm.Invalidate("p-1")
	_, _ = cm.Get(context.Background(), "p-1")

	if calls != 2 {
		t.Errorf("expected 2 DB calls after invalidation, got %d", calls)
	}
}

func TestCredentialManager_ConcurrentAccess(t *testing.T) {
	store := &mockProviderStore{
		onGetProvider: func(_ context.Context, _ string) (*models.Provider, error) {
			creds := &provider.Credentials{AccessToken: "token"}
			return &models.Provider{Credentials: credJSON(t, creds)}, nil
		},
	}

	cm := NewCredentialManager(store)
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = cm.Get(context.Background(), "p-1")
		}()
	}
	wg.Wait()
}
