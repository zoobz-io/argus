package provider

import (
	"fmt"
	"sync"
)

// Registry maps provider types to their implementations.
// Thread-safe for concurrent access from the connector service.
type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewRegistry creates an empty provider registry.
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider implementation for the given type.
// Panics if the type is already registered.
func (r *Registry) Register(p Provider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t := p.Type()
	if _, exists := r.providers[t]; exists {
		panic(fmt.Sprintf("provider type %q already registered", t))
	}
	r.providers[t] = p
}

// Get returns the provider implementation for the given type.
// Returns an error if the type is not registered.
func (r *Registry) Get(providerType string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	p, ok := r.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("unknown provider type: %s", providerType)
	}
	return p, nil
}

// Types returns all registered provider type identifiers.
func (r *Registry) Types() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	types := make([]string, 0, len(r.providers))
	for t := range r.providers {
		types = append(types, t)
	}
	return types
}
