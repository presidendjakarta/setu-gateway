package ratelimiter

import (
	"sync"

	"github.com/presidendjakarta/setu-gateway/internal/logger"
	"github.com/presidendjakarta/setu-gateway/pkg/types"
)

// Manager manages multiple rate limiters
type Manager struct {
	mu        sync.RWMutex
	limiters  map[string]*TokenBucket
	logger    *logger.Logger
}

// NewManager creates a new rate limiter manager
func NewManager(log *logger.Logger) *Manager {
	return &Manager{
		limiters: make(map[string]*TokenBucket),
		logger:   log,
	}
}

// CreateLimiter creates a new rate limiter
func (m *Manager) CreateLimiter(id string, capacity int, refillRate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.limiters[id] = NewTokenBucket(capacity, refillRate)
	m.logger.Infow("Rate limiter created",
		"id", id,
		"capacity", capacity,
		"refill_rate", refillRate,
	)
}

// GetLimiter returns a rate limiter by ID
func (m *Manager) GetLimiter(id string) (*TokenBucket, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	limiter, exists := m.limiters[id]
	return limiter, exists
}

// RemoveLimiter removes a rate limiter
func (m *Manager) RemoveLimiter(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.limiters, id)
	m.logger.Infow("Rate limiter removed", "id", id)
}

// UpdateLimiter updates rate limiter configuration
func (m *Manager) UpdateLimiter(id string, capacity int, refillRate float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.limiters[id]; exists {
		m.limiters[id] = NewTokenBucket(capacity, refillRate)
		m.logger.Infow("Rate limiter updated",
			"id", id,
			"capacity", capacity,
			"refill_rate", refillRate,
		)
	}
}

// CheckRateLimit checks if request is within rate limit
func (m *Manager) CheckRateLimit(limiterID string) bool {
	m.mu.RLock()
	limiter, exists := m.limiters[limiterID]
	m.mu.RUnlock()

	if !exists {
		// No rate limit configured, allow request
		return true
	}

	return limiter.Allow()
}

// FromConfig creates rate limiters from configuration
func (m *Manager) FromConfig(configs []types.RateLimitConfig) {
	for _, cfg := range configs {
		refillRate := float64(cfg.Requests) / cfg.Window.Seconds()
		m.CreateLimiter(cfg.ID, cfg.Requests, refillRate)
	}
}

// ListLimiters returns all limiter IDs
func (m *Manager) ListLimiters() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ids := make([]string, 0, len(m.limiters))
	for id := range m.limiters {
		ids = append(ids, id)
	}

	return ids
}

// Clear removes all rate limiters
func (m *Manager) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.limiters = make(map[string]*TokenBucket)
	m.logger.Info("All rate limiters cleared")
}

// Stats returns rate limiter statistics
func (m *Manager) Stats() map[string]map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]map[string]interface{})
	for id, limiter := range m.limiters {
		stats[id] = map[string]interface{}{
			"capacity":    limiter.Capacity(),
			"tokens":      limiter.Tokens(),
			"available":   limiter.Tokens() > 0,
		}
	}

	return stats
}
