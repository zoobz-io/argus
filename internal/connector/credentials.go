// Package connector provides the credential manager and sync infrastructure
// for the connector service.
package connector

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/zoobz-io/argus/models"
	"github.com/zoobz-io/argus/provider"
)

// ProviderStore defines the database operations needed by the credential manager.
type ProviderStore interface {
	GetProvider(ctx context.Context, id string) (*models.Provider, error)
	UpdateProviderCredentials(ctx context.Context, id, credentials string) error
}

// CredentialManager loads, caches, and refreshes provider credentials.
// Thread-safe: each provider ID has its own mutex to prevent concurrent
// token refresh races.
type CredentialManager struct {
	store  ProviderStore
	cache  map[string]*provider.Credentials
	locks  map[string]*sync.Mutex
	mu     sync.RWMutex
	lockMu sync.Mutex
}

// NewCredentialManager creates a credential manager backed by the given store.
func NewCredentialManager(store ProviderStore) *CredentialManager {
	return &CredentialManager{
		store: store,
		cache: make(map[string]*provider.Credentials),
		locks: make(map[string]*sync.Mutex),
	}
}

// Get returns credentials for the given provider ID.
// Returns cached credentials if available, otherwise loads from the database.
func (cm *CredentialManager) Get(ctx context.Context, providerID string) (*provider.Credentials, error) {
	cm.mu.RLock()
	if creds, ok := cm.cache[providerID]; ok {
		cm.mu.RUnlock()
		return creds, nil
	}
	cm.mu.RUnlock()

	return cm.load(ctx, providerID)
}

// Update persists updated credentials to the database and cache.
// Called when a provider returns refreshed tokens after a data operation.
func (cm *CredentialManager) Update(ctx context.Context, providerID string, creds *provider.Credentials) error {
	if creds == nil {
		return nil
	}

	lock := cm.providerLock(providerID)
	lock.Lock()
	defer lock.Unlock()

	data, err := json.Marshal(creds)
	if err != nil {
		return fmt.Errorf("marshaling credentials: %w", err)
	}

	if err := cm.store.UpdateProviderCredentials(ctx, providerID, string(data)); err != nil {
		return fmt.Errorf("persisting credentials: %w", err)
	}

	cm.mu.Lock()
	cm.cache[providerID] = creds
	cm.mu.Unlock()

	return nil
}

// Invalidate removes cached credentials for a provider, forcing a reload on next Get.
func (cm *CredentialManager) Invalidate(providerID string) {
	cm.mu.Lock()
	delete(cm.cache, providerID)
	cm.mu.Unlock()
}

// load fetches credentials from the database and caches them.
func (cm *CredentialManager) load(ctx context.Context, providerID string) (*provider.Credentials, error) {
	lock := cm.providerLock(providerID)
	lock.Lock()
	defer lock.Unlock()

	// Double-check after acquiring lock.
	cm.mu.RLock()
	if creds, ok := cm.cache[providerID]; ok {
		cm.mu.RUnlock()
		return creds, nil
	}
	cm.mu.RUnlock()

	p, err := cm.store.GetProvider(ctx, providerID)
	if err != nil {
		return nil, fmt.Errorf("loading provider %s: %w", providerID, err)
	}

	var creds provider.Credentials
	if err := json.Unmarshal([]byte(p.Credentials), &creds); err != nil {
		return nil, fmt.Errorf("deserializing credentials for provider %s: %w", providerID, err)
	}

	cm.mu.Lock()
	cm.cache[providerID] = &creds
	cm.mu.Unlock()

	return &creds, nil
}

// providerLock returns a per-provider mutex, creating one if needed.
func (cm *CredentialManager) providerLock(providerID string) *sync.Mutex {
	cm.lockMu.Lock()
	defer cm.lockMu.Unlock()
	if _, ok := cm.locks[providerID]; !ok {
		cm.locks[providerID] = &sync.Mutex{}
	}
	return cm.locks[providerID]
}
