package auth

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Manager handles authentication with multiple providers
type Manager struct {
	mu        sync.RWMutex
	providers map[types.AuthProviderType]AuthProvider
	logger    *logger.Logger
	cache     *AuthCache
}

// NewManager creates a new authentication manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		providers: make(map[types.AuthProviderType]AuthProvider),
		logger:    log,
		cache:     NewAuthCache(),
	}
}

// RegisterProvider registers an authentication provider
func (m *Manager) RegisterProvider(provider AuthProvider) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.providers[provider.Type()] = provider
	m.logger.Infow("Auth provider registered",
		"provider", provider.Name(),
		"type", provider.Type(),
	)

	return nil
}

// GetProvider returns a provider by type
func (m *Manager) GetProvider(authType types.AuthProviderType) (AuthProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	provider, exists := m.providers[authType]
	if !exists {
		return nil, fmt.Errorf("auth provider not found: %s", authType)
	}

	return provider, nil
}

// Authenticate authenticates a request using the auth chain
func (m *Manager) Authenticate(ctx context.Context, r *http.Request, route *types.Route) (*types.AuthResult, error) {
	// If no auth chain, allow request
	if len(route.AuthChain) == 0 {
		return &types.AuthResult{
			Authenticated: true,
			ProviderID:    "none",
			Subject:       "anonymous",
		}, nil
	}

	// Try each provider in the chain
	for _, providerName := range route.AuthChain {
		// Find provider by name
		provider, err := m.findProviderByName(providerName)
		if err != nil {
			m.logger.Warnw("Auth provider not found",
				"provider", providerName,
				"error", err,
			)
			continue
		}

		// Try authentication
		result, err := m.authenticateWithProvider(ctx, r, provider)
		if err != nil {
			m.logger.Warnw("Authentication failed",
				"provider", provider.Name(),
				"error", err,
			)
			// Continue to next provider in chain
			continue
		}

		// Authentication successful
		m.logger.Debugw("Authentication successful",
			"provider", provider.Name(),
			"subject", result.Subject,
		)

		return result, nil
	}

	// All providers failed
	return nil, types.NewGatewayError(
		types.ErrCodeAuthFailed,
		"All authentication providers failed",
		http.StatusUnauthorized,
	)
}

// authenticateWithProvider tries to authenticate with a specific provider
func (m *Manager) authenticateWithProvider(ctx context.Context, r *http.Request, provider AuthProvider) (*types.AuthResult, error) {
	// Check cache first
	cacheKey := m.buildCacheKey(r, provider)
	if result, found := m.cache.Get(cacheKey); found {
		m.logger.Debugw("Auth cache hit",
			"provider", provider.Name(),
		)
		return result, nil
	}

	// Get provider config (from database in production)
	config := types.AuthConfig{
		// TODO: Load from database
	}

	// Authenticate
	result, err := provider.Authenticate(ctx, r, config)
	if err != nil {
		return nil, err
	}

	// Cache result
	m.cache.Set(cacheKey, result)

	return result, nil
}

// findProviderByName finds provider by name
func (m *Manager) findProviderByName(name string) (AuthProvider, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Search through all providers
	for _, provider := range m.providers {
		if provider.Name() == name {
			return provider, nil
		}
	}

	return nil, fmt.Errorf("provider not found: %s", name)
}

// buildCacheKey builds cache key for auth result
func (m *Manager) buildCacheKey(r *http.Request, provider AuthProvider) string {
	// Use request fingerprint + provider name
	// In production, use token/key as cache key
	return fmt.Sprintf("%s:%s:%s",
		provider.Type(),
		r.RemoteAddr,
		r.URL.Path,
	)
}

// ListProviders returns all registered providers
func (m *Manager) ListProviders() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	providers := make([]string, 0, len(m.providers))
	for _, provider := range m.providers {
		providers = append(providers, provider.Name())
	}

	return providers
}

// AuthCache caches authentication results
type AuthCache struct {
	mu    sync.RWMutex
	items map[string]*cacheItem
}

type cacheItem struct {
	result    *types.AuthResult
	expiresAt int64
}

// NewAuthCache creates a new auth cache
func NewAuthCache() *AuthCache {
	return &AuthCache{
		items: make(map[string]*cacheItem),
	}
}

// Get gets cached auth result
func (c *AuthCache) Get(key string) (*types.AuthResult, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	// TODO: Check expiration

	return item.result, true
}

// Set caches auth result
func (c *AuthCache) Set(key string, result *types.AuthResult) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// TODO: Add expiration
	c.items[key] = &cacheItem{
		result: result,
	}
}

// Clear clears the cache
func (c *AuthCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]*cacheItem)
}
